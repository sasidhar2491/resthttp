package resthttp

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type RestHttpError struct {
	HttpStatus int
	HttpReason string
	Msg        string
	Code       string
}

func NewRestHttpError(httpStatus int, httpReason string, msg string, code string) *RestHttpError {
	return &RestHttpError{
		HttpStatus: httpStatus,
		HttpReason: httpReason,
		Msg:        msg,
		Code:       code,
	}
}

func (e *RestHttpError) Error() string {
	if e.Msg != "" {
		return fmt.Sprintf("%d %s: %s", e.HttpStatus, e.HttpReason, e.Msg)
	}
	return fmt.Sprintf("%d %s", e.HttpStatus, e.HttpReason)
}

func (e *RestHttpError) Status() int {
	return e.HttpStatus
}

type ConnectionError struct {
	Msg       string
	ErrorCode int
	Detail    string
}

func NewConnectionError(message string, code int, detail string) *ConnectionError {
	return &ConnectionError{
		Msg:       message,
		ErrorCode: code,
		Detail:    detail,
	}
}

func (e *ConnectionError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s: %s", e.Msg[:len(e.Msg)-1], e.Detail)
	}
	return e.Msg[:len(e.Msg)-1]
}

func (e *ConnectionError) Code() int {
	return e.ErrorCode
}

type RestHttp struct {
	BaseURL     string
	BaseHeaders http.Header
	User        string
	Password    string
	VerifySSL   bool
	DebugPrint  bool
	Timeout     time.Duration
}

func NewRestHttp(baseURL string, user string, password string, sslVerify bool, debugPrint bool, timeout time.Duration) *RestHttp {
	baseURL = strings.TrimRight(baseURL, "/")

	headers := make(http.Header)
	headers.Set("Accept", "application/json")

	if user != "" && password != "" {
		auth := user + ":" + password
		authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
		headers.Set("Authorization", authHeader)
	}

	return &RestHttp{
		BaseURL:     baseURL,
		BaseHeaders: headers,
		User:        user,
		Password:    password,
		VerifySSL:   sslVerify,
		DebugPrint:  debugPrint,
		Timeout:     timeout,
	}
}

func (r *RestHttp) MakeURL(container string, resource string, queryItems url.Values) string {
	parts := []string{r.BaseURL}

	if container != "" {
		parts = append(parts, strings.Trim(container, "/"))
	}

	if resource != "" {
		parts = append(parts, resource)
	} else {
		parts = append(parts, "")
	}

	urlStr := strings.Join(parts, "/")

	if queryItems != nil {
		urlStr += "?" + queryItems.Encode()
	}

	return urlStr
}

func (r *RestHttp) HeadRequest(container string, resource string) (int, error) {
	url := r.MakeURL(container, resource, nil)
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return 0, err
	}

	r.setHeaders(req)

	client := r.createHttpClient()
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}

	if r.DebugPrint {
		r.printRequest("HEAD", resp.Request.URL.String(), req.Header, nil)
	}

	return resp.StatusCode, nil
}

func (r *RestHttp) GetRequest(container string, resource string, queryItems url.Values, accept string, toLower bool) ([]byte, error) {
	url := r.MakeURL(container, resource, queryItems)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if accept != "" {
		req.Header.Set("Accept", accept)
	}

	r.setHeaders(req)

	client := r.createHttpClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if r.DebugPrint {
		r.printRequest("GET", resp.Request.URL.String(), req.Header, nil)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
func (r *RestHttp) setHeaders(req *http.Request) {
	for key, values := range r.BaseHeaders {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
}

func (r *RestHttp) createHttpClient() *http.Client {
	client := &http.Client{
		Timeout: r.Timeout,
	}

	if !r.VerifySSL {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return client
}

func (r *RestHttp) printRequest(method string, url string, headers http.Header, body []byte) {
	fmt.Println("Request:")
	fmt.Println("Method:", method)
	fmt.Println("URL:", url)
	fmt.Println("Headers:")
	for key, values := range headers {
		for _, value := range values {
			fmt.Printf("%s: %s\n", key, value)
		}
	}
	fmt.Println("Body:", string(body))
}

// ...

func (r *RestHttp) PostRequest(container string, resource string, params url.Values, accept string) ([]byte, error) {
	url := r.MakeURL(container, resource, nil)
	req, err := http.NewRequest("POST", url, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if accept != "" {
		req.Header.Set("Accept", accept)
	}

	r.setHeaders(req)

	client := r.createHttpClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if r.DebugPrint {
		r.printRequest("POST", resp.Request.URL.String(), req.Header, []byte(params.Encode()))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (r *RestHttp) PutRequest(container string, resource string, params url.Values, accept string) ([]byte, error) {
	url := r.MakeURL(container, resource, nil)
	req, err := http.NewRequest("PUT", url, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if accept != "" {
		req.Header.Set("Accept", accept)
	}

	r.setHeaders(req)

	client := r.createHttpClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if r.DebugPrint {
		r.printRequest("PUT", resp.Request.URL.String(), req.Header, []byte(params.Encode()))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (r *RestHttp) DeleteRequest(container string, resource string, queryItems url.Values, accept string) ([]byte, error) {
	url := r.MakeURL(container, resource, queryItems)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}

	if accept != "" {
		req.Header.Set("Accept", accept)
	}

	r.setHeaders(req)

	client := r.createHttpClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if r.DebugPrint {
		r.printRequest("DELETE", resp.Request.URL.String(), req.Header, nil)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (r *RestHttp) DownloadFile(container string, resource string, savePath string, accept string, queryItems url.Values) error {
	resource = strings.ReplaceAll(resource, "\\", "/")
	url := r.MakeURL(container, resource, queryItems)
	if savePath == "" {
		savePath = strings.Split(resource, "/")[len(strings.Split(resource, "/"))-1]
	}

	headers := r.BaseHeaders.Clone()
	if accept != "" {
		headers.Set("Accept", accept)
	}

	if queryItems != nil && len(queryItems) > 0 {
		url += "?" + queryItems.Encode()
		queryItems = nil
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	r.setHeaders(req)

	client := r.createHttpClient()
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return NewRestHttpError(resp.StatusCode, resp.Status, "", "")
	}

	fileSizeDl := 0
	file, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("could not create file: %s", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("could not download file: %s", err)
	}

	if r.DebugPrint {
		fmt.Printf("===> downloaded %d bytes to %s\n", fileSizeDl, savePath)
	}

	return nil
}

func (r *RestHttp) UploadFile(container string, resource string, params url.Values, contentType string, file *os.File) ([]byte, error) {
	url := r.MakeURL(container, resource, nil)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}

	// Add additional parameters
	for key, values := range params {
		for _, value := range values {
			_ = writer.WriteField(key, value)
		}
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.setHeaders(req)

	client := r.createHttpClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if r.DebugPrint {
		r.printRequest("POST", resp.Request.URL.String(), req.Header, nil)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}
func (r *RestHttp) handleResponse(resp *http.Response) error {
	if resp.StatusCode >= 300 {
		return NewRestHttpError(resp.StatusCode, resp.Status, "", "")
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (r *RestHttp) UploadFileMP(container string, srcFilePath string, dstName string, contentType string) ([]byte, error) {
	if !fileExists(srcFilePath) {
		return nil, fmt.Errorf("file not found: %s", srcFilePath)
	}

	if dstName == "" {
		dstName = filepath.Base(srcFilePath)
	}

	if contentType == "" {
		contentType = "application/octet.stream"
	}

	url := r.MakeURL(container, "", nil)
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	file, err := os.Open(srcFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", dstName)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}

	contentType = writer.FormDataContentType()
	writer.Close()

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	r.setHeaders(req)

	client := r.createHttpClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if r.DebugPrint {
		r.printRequest("POST", resp.Request.URL.String(), req.Header, nil)
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return responseBody, nil
}

func (r *RestHttp) UploadFiles(container string, srcDstMap map[string]string, contentType string) ([]byte, error) {
	if contentType == "" {
		contentType = "application/octet.stream"
	}

	url := r.MakeURL(container, "", nil)
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	var fileCloseFuncs []func()

	for srcPath, dstName := range srcDstMap {
		if dstName == "" {
			dstName = filepath.Base(srcPath)
		}

		file, err := os.Open(srcPath)
		if err != nil {
			return nil, err
		}

		part, err := writer.CreateFormFile("files", dstName)
		if err != nil {
			file.Close()
			return nil, err
		}

		_, err = io.Copy(part, file)
		if err != nil {
			file.Close()
			return nil, err
		}

		fileCloseFuncs = append(fileCloseFuncs, func() { file.Close() })
	}

	contentType = writer.FormDataContentType()
	writer.Close()

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		for _, fileCloseFunc := range fileCloseFuncs {
			fileCloseFunc()
		}
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	r.setHeaders(req)

	client := r.createHttpClient()
	resp, err := client.Do(req)
	if err != nil {
		for _, fileCloseFunc := range fileCloseFuncs {
			fileCloseFunc()
		}
		return nil, err
	}
	defer resp.Body.Close()

	if r.DebugPrint {
		r.printRequest("POST", resp.Request.URL.String(), req.Header, nil)
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		for _, fileCloseFunc := range fileCloseFuncs {
			fileCloseFunc()
		}
		return nil, err
	}

	return responseBody, nil
}
