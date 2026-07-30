package quickbooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
	"time"
)

type ContentType string

const (
	AI   ContentType = "application/postscript"
	CSV  ContentType = "text/csv"
	DOC  ContentType = "application/msword"
	DOCX ContentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	EPS  ContentType = "application/postscript"
	GIF  ContentType = "image/gif"
	JPEG ContentType = "image/jpeg"
	JPG  ContentType = "image/jpg"
	ODS  ContentType = "application/vnd.oasis.opendocument.spreadsheet"
	PDF  ContentType = "application/pdf"
	PNG  ContentType = "image/png"
	RTF  ContentType = "text/rtf"
	TIF  ContentType = "image/tif"
	TXT  ContentType = "text/plain"
	XLS  ContentType = "application/vnd/ms-excel"
	XLSX ContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	XML  ContentType = "text/xml"
)

type Attachable struct {
	Id                       string          `json:"Id,omitempty"`
	SyncToken                string          `json:",omitempty"`
	FileName                 string          `json:",omitempty"`
	Note                     string          `json:",omitempty"`
	Category                 string          `json:",omitempty"`
	ContentType              ContentType     `json:",omitempty"`
	PlaceName                string          `json:",omitempty"`
	AttachableRef            []AttachableRef `json:",omitempty"`
	Long                     string          `json:",omitempty"`
	Tag                      string          `json:",omitempty"`
	Lat                      string          `json:",omitempty"`
	MetaData                 MetaData        `json:",omitempty"`
	FileAccessUri            string          `json:",omitempty"`
	Size                     json.Number     `json:",omitempty"`
	ThumbnailFileAccessUri   string          `json:",omitempty"`
	TempDownloadUri          string          `json:",omitempty"`
	ThumbnailTempDownloadUri string          `json:",omitempty"`
}

type AttachableRef struct {
	IncludeOnSend bool   `json:",omitempty"`
	LineInfo      string `json:",omitempty"`
	NoRefOnly     bool   `json:",omitempty"`
	// CustomField[0..n]
	Inactive  bool          `json:",omitempty"`
	EntityRef ReferenceType `json:",omitempty"`
}

// CreateAttachable creates the given Attachable on the QuickBooks server,
// returning the resulting Attachable object.
func (c *Client) CreateAttachable(attachable *Attachable) (*Attachable, error) {
	var resp struct {
		Attachable Attachable
		Time       Date
	}

	if err := c.post("attachable", attachable, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Attachable, nil
}

// DeleteAttachable deletes the attachable
func (c *Client) DeleteAttachable(attachable *Attachable) error {
	if attachable.Id == "" || attachable.SyncToken == "" {
		return fmt.Errorf("%w: missing id/sync token", ErrMissingID)
	}

	return c.post("attachable", attachable, nil, map[string]string{"operation": "delete"})
}

// DownloadAttachable downloads the attachable
func (c *Client) DownloadAttachable(id string) (string, error) {
	endpointUrl := *c.endpoint
	endpointUrl.Path += "download/" + id

	urlValues := url.Values{}
	urlValues.Add("minorversion", c.minorVersion)
	endpointUrl.RawQuery = urlValues.Encode()

	req, err := http.NewRequest("GET", endpointUrl.String(), nil)
	if err != nil {
		return "", err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", parseFailure(resp)
	}

	downloadUrl, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(downloadUrl), err
}

// DownloadAttachableContent downloads the attachable file content, streaming it to the provided writer.
// Returns the content type and number of bytes written.
// Uses a plain HTTP client for the actual file download since the temporary download URL
// has its own auth and does not accept OAuth headers.
func (c *Client) DownloadAttachableContent(id string, w io.Writer) (string, int64, error) {
	downloadUrl, err := c.DownloadAttachable(id)
	if err != nil {
		return "", 0, fmt.Errorf("get download URL: %w", err)
	}

	if downloadUrl == "" {
		return "", 0, ErrNoDownloadURL
	}

	// Use a plain HTTP client — the temp download URL has embedded auth
	// and will reject OAuth headers with 403
	plainClient := &http.Client{Timeout: 60 * time.Second}
	resp, err := plainClient.Get(downloadUrl)
	if err != nil {
		return "", 0, fmt.Errorf("fetch file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, &HTTPError{StatusCode: resp.StatusCode, Message: "fetch file"}
	}

	contentType := resp.Header.Get("Content-Type")

	n, err := io.Copy(w, resp.Body)
	if err != nil {
		return contentType, n, fmt.Errorf("stream file: %w", err)
	}

	return contentType, n, nil
}

// EntityName returns the QuickBooks entity name for Attachable.
func (Attachable) EntityName() string { return "Attachable" }

func (v Attachable) entityId() string { return v.Id }

func (v Attachable) entityCreateTime() Date { return v.MetaData.CreateTime }

// AttachableFilter selects which Attachables a scan returns.
//
// The values are opaque on purpose. An earlier API took a caller-supplied QBO
// SQL string, which meant the library could not paginate it (see the comment at
// the top of pagination.go) and put query syntax in every caller. Callers now
// name a filter and the library owns the clause behind it. Add a constant here
// to support a new one.
type AttachableFilter int

const (
	// AllAttachables places no restriction on the scan. It is the zero
	// value, so the empty filter means "everything".
	AllAttachables AttachableFilter = iota

	// InvoiceAttachables restricts the scan to attachments linked to invoices.
	InvoiceAttachables

	// PaymentAttachables restricts the scan to attachments linked to payments.
	PaymentAttachables
)

// predicate returns the QBO SQL fragment for f, or an error if f is not one of
// the defined filters. An unknown value fails the scan rather than silently
// widening it to everything.
func (f AttachableFilter) predicate() (string, error) {
	switch f {
	case AllAttachables:
		return "", nil
	case InvoiceAttachables:
		return "AttachableRef.EntityRef.Type = 'Invoice'", nil
	case PaymentAttachables:
		return "AttachableRef.EntityRef.Type = 'Payment'", nil
	}

	return "", fmt.Errorf("unknown attachable filter %d", int(f))
}

// String implements fmt.Stringer so a filter reads sensibly in logs.
func (f AttachableFilter) String() string {
	switch f {
	case AllAttachables:
		return "all"
	case InvoiceAttachables:
		return "invoice"
	case PaymentAttachables:
		return "payment"
	}

	return "AttachableFilter(" + strconv.Itoa(int(f)) + ")"
}

// FindAttachables gets the full list of Attachables matching filter. Pass
// AllAttachables for every attachment in the account.
func (c *Client) FindAttachables(filter AttachableFilter) ([]Attachable, error) {
	predicate, err := filter.predicate()
	if err != nil {
		return nil, err
	}

	return findAllWhere[Attachable](c, predicate)
}

// FindAttachableById finds the attachable by the given id
func (c *Client) FindAttachableById(id string) (*Attachable, error) {
	var resp struct {
		Attachable Attachable
		Time       Date
	}

	if err := c.get("attachable/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Attachable, nil
}

// UpdateAttachable updates the attachable
func (c *Client) UpdateAttachable(attachable *Attachable) (*Attachable, error) {
	if attachable.Id == "" {
		return nil, fmt.Errorf("%w: missing attachable id", ErrMissingID)
	}

	existingAttachable, err := c.FindAttachableById(attachable.Id)
	if err != nil {
		return nil, err
	}

	attachable.SyncToken = existingAttachable.SyncToken

	payload := struct {
		*Attachable
		Sparse bool `json:"sparse"`
	}{
		Attachable: attachable,
		Sparse:     true,
	}

	var attachableData struct {
		Attachable Attachable
		Time       Date
	}

	if err = c.post("attachable", payload, &attachableData, nil); err != nil {
		return nil, err
	}

	return &attachableData.Attachable, err
}

// UploadAttachable uploads the attachable
func (c *Client) UploadAttachable(attachable *Attachable, data io.Reader) (*Attachable, error) {
	endpointUrl := *c.endpoint
	endpointUrl.Path += "upload"

	urlValues := url.Values{}
	urlValues.Add("minorversion", c.minorVersion)
	endpointUrl.RawQuery = urlValues.Encode()

	var buffer bytes.Buffer
	mWriter := multipart.NewWriter(&buffer)

	// Add file metadata
	metadataHeader := make(textproto.MIMEHeader)
	metadataHeader.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file_metadata_01", "attachment.json"))
	metadataHeader.Set("Content-Type", "application/json")

	metadataContent, err := mWriter.CreatePart(metadataHeader)
	if err != nil {
		return nil, err
	}

	j, err := json.Marshal(attachable)
	if err != nil {
		return nil, err
	}

	if _, err = metadataContent.Write(j); err != nil {
		return nil, err
	}

	// Add file content
	fileHeader := make(textproto.MIMEHeader)
	fileHeader.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file_content_01", attachable.FileName))
	fileHeader.Set("Content-Type", string(attachable.ContentType))

	fileContent, err := mWriter.CreatePart(fileHeader)
	if err != nil {
		return nil, err
	}

	if _, err = io.Copy(fileContent, data); err != nil {
		return nil, err
	}

	mWriter.Close()

	req, err := http.NewRequest("POST", endpointUrl.String(), &buffer)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", mWriter.FormDataContentType())
	req.Header.Add("Accept", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, parseFailure(resp)
	}

	var r struct {
		AttachableResponse []struct {
			Attachable Attachable
		}
		Time Date
	}

	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}

	return &r.AttachableResponse[0].Attachable, nil
}
