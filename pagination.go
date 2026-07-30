package quickbooks

import (
	"fmt"
	"time"
)

// Keyset pagination over MetaData.CreateTime.
//
// The obvious way to walk an entity is STARTPOSITION/MAXRESULTS with ORDERBY
// Id, and that is what this library used to do. Two things make it unsound:
//
//   - Offsets are positional. A row deleted mid-scan shifts every later row
//     down one, so the row that was at the next page's first position is never
//     returned. Nothing reports this; the scan just comes back short.
//   - Intuit has announced that Id is no longer a sortable field:
//     https://medium.com/intuitdev/upcoming-changes-to-apis-and-tools-that-may-impact-your-application-df4680ae3535
//     Without a sort, QuickBooks returns rows most-recently-modified first
//     (verified against a live company), which is the worst possible order
//     for offsets: editing any row moves it to position 1 and shifts
//     everything after it.
//
// Paging by a cursor instead of a position fixes both: each request asks for
// rows at or after a value rather than at an index, so shifting rows cannot
// cause a skip. MetaData.CreateTime is used as the key because it is
// system-assigned and immutable (TxnDate is user-editable and can move behind
// a cursor already passed), it exists on every entity (TxnDate does not --
// Customer and Item have no such field), and it increases with creation order.
//
// CreateTime is not unique: records created in bulk can share a second. The
// predicate is therefore ≥ rather than >, and rows already collected are
// skipped by Id. That potentially duplicates rows on subsequent pages in
// exchange for never dropping a member of a tie group. However, if ≥ pagesize
// records are created with the same creation timestamp, it means the page is
// entirely records we've already received. Such a page stalls the cursor's
// forward movement. We explicitly check for and error this case. That's not
// ideal. We prefer an alternative solution, but couldn't identify one. We
// think this is the best possible implementation at the time of writing based
// on the API constraints. The rationale follows from the observed create time
// precision of 1 second, and default pagesize = 1,000. A batch create of ≥ 1,000
// is wildly unlikely for our use case, so this is an acceptable tradeoff
// for the other correctness guarantees. If you use this for high volume
// datasets, or you need to shrink pagesize for some reason, YMMV.

type createTimeCursor struct {
	value string
	seen  map[string]bool
}

func newCreateTimeCursor() *createTimeCursor {
	return &createTimeCursor{seen: map[string]bool{}}
}

// where returns the predicate selecting the next page, or "" for the first.
func (cur *createTimeCursor) where() string {
	if cur.value == "" {
		return ""
	}
	return " WHERE MetaData.CreateTime >= '" + cur.value + "'"
}

// collect reports whether id is new, recording it when so. Rows that are not
// new are the tie-group overlap produced by the >= predicate.
func (cur *createTimeCursor) collect(id string) bool {
	if cur.seen[id] {
		return false
	}
	cur.seen[id] = true
	return true
}

// advance moves the cursor to the CreateTime of the last row of the page just
// processed.
//
// The value is interpolated into a query literal, so it is validated rather
// than trusted. It arrives from QuickBooks rather than from a caller, but a
// malformed value would either change which rows the scan sees or break out of
// the quoted literal, and neither failure would be obvious.
func (cur *createTimeCursor) advance(d Date) error {
	s := d.String()
	if s == "" {
		return fmt.Errorf("cannot page by MetaData.CreateTime: row has no CreateTime")
	}

	if _, err := time.Parse(time.RFC3339, s); err != nil {
		return fmt.Errorf("cannot page by MetaData.CreateTime %q: not an RFC3339 timestamp: %w", s, err)
	}

	cur.value = s

	return nil
}

// stalled builds the error for a full page that contained nothing new, which
// means a single CreateTime value spans at least a whole page and the cursor
// cannot move past it.
func (cur *createTimeCursor) stalled(entity string, pageSize int) error {
	return fmt.Errorf("%s pagination stalled at MetaData.CreateTime %q: more than %d rows share that timestamp",
		entity, cur.value, pageSize)
}

// countResponse is the envelope QuickBooks returns for a SELECT COUNT(*)
// query. The shape does not vary by entity.
//
// This is the only place a TotalCount field appears. The per-entity paged
// response types deliberately omit it so that a count can never reach the
// pagination logic -- it is a check applied after a scan, never an input to it.
type countResponse struct {
	QueryResponse struct {
		TotalCount int
	}
}

// countEntity asks QuickBooks how many rows an entity holds.
func (c *Client) countEntity(entity string) (int, error) {
	var count countResponse

	if err := c.query("SELECT COUNT(*) FROM "+entity, &count); err != nil {
		return 0, err
	}

	return count.QueryResponse.TotalCount, nil
}

// verifyScanComplete compares what a scan collected against the row count
// QuickBooks reported before the scan began. Fewer rows than reported is an
// error; more is not. The asymmetry is deliberate, and the reasoning is spelled
// out here because it rests on assumptions that may not survive contact with
// production.
//
// Why a shortfall must fail. Callers of the Find* helpers diff the returned
// slice against their own stored copy and treat anything absent from it as
// deleted upstream. A scan that quietly returns 900 of 1000 rows therefore does
// not read as "900 rows"; it reads as "100 rows were deleted". Returning an
// error instead of a short slice is what keeps a fetch problem from being
// misread as an instruction to delete data.
//
// Why a surplus is safe. The count is taken before the scan, so a row created
// while the scan runs is not counted -- but it is collected, because under
// CreateTime ASC a new row sorts after everything already fetched. Surplus
// therefore means "the world grew while we looked at it", and no row is
// missing. Rows are deduplicated by Id during the scan, so a surplus cannot be
// an artifact of re-reading the same row across the >= page boundary; if that
// dedupe is ever removed, this reasoning breaks and a surplus stops being
// meaningful.
//
// Why the check earns its cost. Termination trusts that a short page means "no
// more rows". Whether QuickBooks can return a short page mid-scan for its own
// reasons -- a deletion between page requests, an internal limit, throttling --
// is not something we have enough production experience to rule out. Without
// this comparison, any such page would end a scan silently and the result would
// look complete. A false positive costs one skipped sync and clears itself on
// the next run; a false negative costs real data. Deletions during a scan do
// land in the shortfall direction and will trip this. That is accepted.
//
// Known hole. This compares totals, not identities, so it cannot see a miss and
// a create that cancel out: skip three rows while five are created and the
// count still passes. It is a tripwire, not a proof of completeness. A caller
// that needs certainty has to compare identities against its own records.
//
// If this ever fires spuriously, suspect these in order. First, whether
// COUNT(*) and SELECT * filter identically for that entity -- if COUNT includes
// rows the scan does not return, the shortfall is permanent rather than racy,
// and the finder will fail every time. Invoice and Customer were verified to
// agree exactly against a live company; every entity converted after them needs
// the same confirmation. Second, whether a tie group is being silently dropped
// at a page boundary. Third, whether short-page-means-done still holds.
func verifyScanComplete(entity string, collected, reported int) error {
	if collected >= reported {
		return nil
	}

	return fmt.Errorf("%w: %s scan collected %d of the %d rows QuickBooks reported",
		ErrIncompleteScan, entity, collected, reported)
}
