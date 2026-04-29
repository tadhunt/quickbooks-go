// Copyright (c) 2018, Randy Westlund. All rights reserved.
// This code is under the BSD-2-Clause license.
package quickbooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

// Client is your handle to the QuickBooks API.
type Client struct {
	// Get this from oauth2.NewClient().
	Client *http.Client
	// Set to ProductionEndpoint or SandboxEndpoint.
	endpoint *url.URL
	// The set of quickbooks APIs
	discoveryAPI *DiscoveryAPI
	// The client Id
	clientId string
	// The client Secret
	clientSecret string
	// The minor version of the QB API
	minorVersion string
	// The account Id you're connecting to.
	realmId string
	// Flag set if the limit of 500req/s has been hit (source: https://developer.intuit.com/app/developer/qbo/docs/learn/rest-api-features#limits-and-throttles)
	throttled bool

	debug string

	// companyTZ is the resolved IANA location for the QuickBooks
	// company. Fetched from CompanyInfo.DefaultTimeZone during
	// NewClient and used to anchor bare-date fields like TxnDate.
	companyTZ *time.Location
}

// CompanyTimezone returns the QuickBooks company's configured timezone,
// fetched from CompanyInfo.DefaultTimeZone during NewClient. Non-nil for
// any Client constructed with a bearer token; nil when constructed
// without one (the OAuth setup flow that produces an auth URL).
func (c *Client) CompanyTimezone() *time.Location {
	return c.companyTZ
}

// Time parses d using the company timezone. RFC3339 values ignore the
// company timezone (they carry their own offset); bare YYYY-MM-DD values
// are anchored at midnight in the company timezone.
func (c *Client) Time(d Date) (time.Time, error) {
	return d.In(c.companyTZ)
}

// NewClient initializes a new QuickBooks client for interacting with their Online API
func NewClient(clientId string, clientSecret string, realmId string, isProduction bool, minorVersion string, token *BearerToken, debug string) (c *Client, err error) {
	if minorVersion == "" {
		minorVersion = "70"
	}

	client := Client{
		clientId:     clientId,
		clientSecret: clientSecret,
		minorVersion: minorVersion,
		realmId:      realmId,
		throttled:    false,
		debug:        debug,
	}

	if isProduction {
		client.endpoint, err = url.Parse(ProductionEndpoint.String() + "/v3/company/" + realmId + "/")
		if err != nil {
			return nil, fmt.Errorf("failed to parse API endpoint: %w", err)
		}

		client.discoveryAPI, err = CallDiscoveryAPI(DiscoveryProductionEndpoint)
		if err != nil {
			return nil, fmt.Errorf("failed to obtain discovery endpoint: %w", err)
		}
	} else {
		client.endpoint, err = url.Parse(SandboxEndpoint.String() + "/v3/company/" + realmId + "/")
		if err != nil {
			return nil, fmt.Errorf("failed to parse API endpoint: %w", err)
		}

		client.discoveryAPI, err = CallDiscoveryAPI(DiscoverySandboxEndpoint)
		if err != nil {
			return nil, fmt.Errorf("failed to obtain discovery endpoint: %w", err)
		}
	}

	if token != nil {
		client.Client = getHttpClient(token)
	}

	// Fetch company timezone so bare-date fields can be anchored
	// correctly. We fail construction if this can't be resolved
	// because silently misinterpreting dates causes hard-to-spot
	// data corruption (e.g. invoices recorded on the wrong day).
	if client.Client != nil {
		info, err := client.FindCompanyInfo()
		if err != nil {
			return nil, fmt.Errorf("fetch company info for timezone: %w", err)
		}
		if info.DefaultTimeZone == "" {
			return nil, fmt.Errorf("company info missing DefaultTimeZone (realm %s)", realmId)
		}
		loc, err := time.LoadLocation(info.DefaultTimeZone)
		if err != nil {
			return nil, fmt.Errorf("load company timezone %q: %w", info.DefaultTimeZone, err)
		}
		client.companyTZ = loc
	}

	return &client, nil
}

// FindAuthorizationUrl compiles the authorization url from the discovery api's auth endpoint.
//
// Example: qbClient.FindAuthorizationUrl("com.intuit.quickbooks.accounting", "security_token", "https://developer.intuit.com/v2/OAuth2Playground/RedirectUrl")
//
// You can find live examples from https://developer.intuit.com/app/developer/playground
func (c *Client) FindAuthorizationUrl(scope string, state string, redirectUri string) (string, error) {
	var authorizationUrl *url.URL

	authorizationUrl, err := url.Parse(c.discoveryAPI.AuthorizationEndpoint)
	if err != nil {
		return "", fmt.Errorf("failed to parse auth endpoint: %w", err)
	}

	urlValues := url.Values{}
	urlValues.Add("client_id", c.clientId)
	urlValues.Add("response_type", "code")
	urlValues.Add("scope", scope)
	urlValues.Add("redirect_uri", redirectUri)
	urlValues.Add("state", state)
	authorizationUrl.RawQuery = urlValues.Encode()

	return authorizationUrl.String(), nil
}

func (c *Client) req(method string, endpoint string, payloadData interface{}, responseObject interface{}, queryParameters map[string]string) error {
	// TODO: possibly just wait until c.throttled is false, and continue the request?
	if c.throttled {
		return ErrRateLimit
	}

	endpointUrl := *c.endpoint
	endpointUrl.Path += endpoint
	urlValues := url.Values{}

	if len(queryParameters) > 0 {
		for param, value := range queryParameters {
			urlValues.Add(param, value)
		}
	}

	urlValues.Set("minorversion", c.minorVersion)
	urlValues.Set("include", "enhancedAllCustomFields")
	endpointUrl.RawQuery = urlValues.Encode()

	var err error
	var marshalledJson []byte

	if payloadData != nil {
		marshalledJson, err = json.Marshal(payloadData)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
	}

	req, err := http.NewRequest(method, endpointUrl.String(), bytes.NewBuffer(marshalledJson))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	c.dumpRequest(req)

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}

	c.dumpResponse(resp)

	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		break
	case http.StatusTooManyRequests:
		c.throttled = true
		go func(c *Client) {
			time.Sleep(1 * time.Minute)
			c.throttled = false
		}(c)
	default:
		return parseFailure(resp)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if responseObject != nil {
		err = json.Unmarshal(data, &responseObject)
		if err != nil {
			return fmt.Errorf("failed to unmarshal response into object: %w", err)
		}
		//		if err = json.NewDecoder(resp.Body).Decode(&responseObject); err != nil {
		//			return fmt.Errorf("failed to unmarshal response into object: %v", err)
		//		}
	}

	return nil
}

func (c *Client) dumpRequest(req *http.Request) {
	if c.debug == "" {
		return
	}

	f, err := os.OpenFile(c.debug, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("open %s: %v", c.debug, err)
		return
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Printf("close %s: %v", c.debug, err)
		}
	}()

	data, err := httputil.DumpRequest(req, true)
	if err != nil {
		log.Printf("dumprequest: %v", err)
		return
	}

	data = append(data, '\n')

	_, err = f.Write(data)
	if err != nil {
		log.Printf("write %s: %v", c.debug, err)
		return
	}
}

func (c *Client) dumpResponse(resp *http.Response) {
	if c.debug == "" {
		return
	}

	f, err := os.OpenFile(c.debug, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("open %s: %v", c.debug, err)
		return
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Printf("close %s: %v", c.debug, err)
		}
	}()

	data, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Printf("dumpresponse: %v", err)
		return
	}

	data = append(data, '\n')

	_, err = f.Write(data)
	if err != nil {
		log.Printf("write %s: %v", c.debug, err)
		return
	}
}

func (c *Client) get(endpoint string, responseObject interface{}, queryParameters map[string]string) error {
	return c.req("GET", endpoint, nil, responseObject, queryParameters)
}

func (c *Client) post(endpoint string, payloadData interface{}, responseObject interface{}, queryParameters map[string]string) error {
	return c.req("POST", endpoint, payloadData, responseObject, queryParameters)
}

// query makes the specified QBO `query` and unmarshals the result into `responseObject`
func (c *Client) query(query string, responseObject interface{}) error {
	return c.get("query", responseObject, map[string]string{"query": query})
}

// Query is a public wrapper around query for ad-hoc QBO SQL (probes, diagnostics).
func (c *Client) Query(qstr string, responseObject interface{}) error {
	return c.query(qstr, responseObject)
}
