package client

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"gitlab.com/dataptive/styx/api"
	"gitlab.com/dataptive/styx/log"

	"github.com/gorilla/schema"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) (c *Client) {

	c = &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}

	return c
}

func (c *Client) ListLogs() (r api.ListLogsResponse, err error) {

	endpoint := c.baseURL + "/logs"

	resp, err := c.httpClient.Get(endpoint)
	if err != nil {
		return r, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = api.ReadError(resp.Body)
		return r, err
	}

	api.ReadResponse(resp.Body, &r)

	return r, nil
}

func (c *Client) CreateLog(name string, config api.LogConfig) (r api.CreateLogResponse, err error) {

	endpoint := c.baseURL + "/logs"

	encoder := schema.NewEncoder()

	logForm := api.CreateLogForm{
		Name:      name,
		LogConfig: &config,
	}
	form := url.Values{}

	err = encoder.Encode(logForm, form)
	if err != nil {
		return r, err
	}

	resp, err := c.httpClient.PostForm(endpoint, form)
	if err != nil {
		return r, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = api.ReadError(resp.Body)
		return r, err
	}

	api.ReadResponse(resp.Body, &r)

	return r, nil
}

func (c *Client) GetLog(name string) (r api.GetLogResponse, err error) {

	endpoint := c.baseURL + "/logs/" + name

	resp, err := c.httpClient.Get(endpoint)
	if err != nil {
		return r, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = api.ReadError(resp.Body)
		return r, err
	}

	api.ReadResponse(resp.Body, &r)

	return r, nil
}

func (c *Client) DeleteLog(name string) (err error) {

	endpoint := c.baseURL + "/logs/" + name

	req, err := http.NewRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = api.ReadError(resp.Body)
		return err
	}

	return nil
}

func (c *Client) BackupLog(name string, w io.Writer) (err error) {

	endpoint := c.baseURL + "/logs/" + name + "/backup"

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = api.ReadError(resp.Body)
		return err
	}

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) RestoreLog(name string, r io.Reader) (err error) {

	endpoint := c.baseURL + "/logs/restore?name=" + name

	resp, err := c.httpClient.Post(endpoint, "application/gzip", r)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = api.ReadError(resp.Body)
		return err
	}

	return nil
}

func (c *Client) WriteRecord(logName string, record log.Record) (r api.WriteRecordResponse, err error) {

	endpoint := c.baseURL + "/logs/" + logName + "/records"

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(record))
	if err != nil {
		return r, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return r, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = api.ReadError(resp.Body)
		return r, err
	}

	api.ReadResponse(resp.Body, &r)

	return r, nil
}

func (c *Client) ReadRecord(logName string, params api.ReadRecordParams) (r log.Record, err error) {

	encoder := schema.NewEncoder()
	queryParams := url.Values{}

	err = encoder.Encode(params, queryParams)
	if err != nil {
		return r, err
	}

	endpoint := c.baseURL + "/logs/" + logName + "/records?" + queryParams.Encode()

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return r, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return r, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = api.ReadError(resp.Body)
		return r, err
	}

	test, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return r, err
	}

	r = log.Record(test)

	return r, nil
}
