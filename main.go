package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	baseURL     = "https://api.packet.net/"
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var (
	token        *string
	projectID    *string
	hostname     *string
	facility     *string
	plan         *string
	ops          *string
	billingCycle *string
)

// Client is HTTP client
type Client struct {
	baseURL string
	token   string
	client  *http.Client
}

// NewClient creates a Client instance
func NewClient(token, apiURL string) *Client {
	return &Client{
		token:   token,
		baseURL: apiURL,
		client:  &http.Client{},
	}
}

// DoRequest performs HTTP request
func (c *Client) DoRequest(url string, method string, request interface{}, response interface{}, raw *string) error {
	var payload io.Reader

	if request != nil {
		data, err := json.Marshal(request)
		if err != nil {
			return err
		}
		payload = bytes.NewBuffer(data)
	}

	r, err := http.NewRequest(method, c.baseURL+url, payload)
	if err != nil {
		return err
	}

	r.Header.Add("X-Auth-Token", c.token)
	r.Header.Add("Content-Type", "application/json")

	resp, err := c.client.Do(r)
	if err != nil {
		return err
	}

	if resp != nil {
		var body []byte
		defer resp.Body.Close()
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if raw != nil {
			*raw = string(body)
		}

		if method != "DELETE" {
			err = json.Unmarshal(body, response)
		}
	}

	return err
}

// DeviceRequest is used to create a Packet device
type DeviceRequest struct {
	Hostname     string   `json:"hostname"`
	Plan         string   `json:"plan"`
	Facility     []string `json:"facility"`
	OS           string   `json:"operating_system"`
	BillingCycle string   `json:"billing_cycle"`
	ProjectID    string   `json:"project_id"`
}

// Device represents a Packet device API instance
type Device struct {
	ID           string                 `json:"id"`
	Hostname     string                 `json:"hostname,omitempty"`
	State        string                 `json:"state,omitempty"`
	Created      string                 `json:"created_at,omitempty"`
	Updated      string                 `json:"updated_at,omitempty"`
	Locked       bool                   `json:"locked,omitempty"`
	BillingCycle string                 `json:"billing_cycle,omitempty"`
	Storage      map[string]interface{} `json:"storage,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	Network      interface{}            `json:"ip_addresses"`
	Volumes      interface{}            `json:"volumes"`
	OS           interface{}            `json:"operating_system,omitempty"`
	Plan         interface{}            `json:"plan,omitempty"`
	Facility     interface{}            `json:"facility,omitempty"`
	Project      interface{}            `json:"project,omitempty"`
}

func main() {
	parseInputParams()

	client := NewClient(*token, baseURL)

	device := createDevice(client)

	if device != nil {
		fmt.Println("Device is ready. Terminating in 10s...")
		time.Sleep(10 * time.Second)
		deleteDevice(device.ID, client)
	}
}

func parseInputParams() {
	rand.Seed(time.Now().UnixNano())

	// generate random name for the device, if not provided
	b := make([]byte, 15)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	name := string(b)

	token = flag.String("token", os.Getenv("PACKET_AUTH_TOKEN"), "Packet API key token")
	projectID = flag.String("prid", os.Getenv("PACKET_PROJECT_ID"), "project ID")

	hostname = flag.String("hostname", name, "Hostname of the server to be deployed")
	facility = flag.String("facility", "ams1", "Datacenter facility code where to deploy device")
	plan = flag.String("plan", "baremetal_0", "Server deployment plan")
	ops = flag.String("os", "centos_7", "Server OS slug")
	billingCycle = flag.String("bilcycle", "hourly", "Billing cycle")

	flag.Parse()

	if strings.TrimSpace(*token) == "" {
		fmt.Println("You must provide Packet API token. Set PACKET_AUTH_TOKEN env variable or provide --token flag.")
		os.Exit(0)
	}

	if strings.TrimSpace(*projectID) == "" {
		fmt.Println("You must provide project ID. Set PACKET_PROJECT_ID env variable or provide --prid flag.")
		os.Exit(0)
	}
}

func createDevice(client *Client) *Device {
	devReq := &DeviceRequest{
		Hostname:     *hostname,
		Facility:     []string{*facility},
		Plan:         *plan,
		OS:           *ops,
		ProjectID:    *projectID,
		BillingCycle: *billingCycle,
	}

	uri := fmt.Sprintf("projects/%s/devices", *projectID)

	device := new(Device)
	// raw response might be usefull for troubleshooting
	rawResponse := new(string)

	err := client.DoRequest(uri, "POST", &devReq, device, rawResponse)

	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	fmt.Println("Provisioning device... please wait")

	device, err = waitUntilReady(device.ID, client)

	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	prettyPrint(device)
	return device
}

func deleteDevice(deviceID string, client *Client) {
	uri := "devices/" + deviceID
	err := client.DoRequest(uri, "DELETE", nil, nil, nil)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("Device %s successfully deleted\n", deviceID)
}

func prettyPrint(in interface{}) {
	res, err := json.MarshalIndent(in, "", "  ")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(string(res))
}

func waitUntilReady(deviceID string, c *Client) (*Device, error) {
	for i := 0; i < 300; i++ {
		time.Sleep(5 * time.Second)
		dev := new(Device)
		err := c.DoRequest("devices/"+deviceID, "GET", nil, dev, nil)
		if err != nil {
			return nil, err
		}
		if dev.State == "active" {
			return dev, nil
		}
	}
	return nil, fmt.Errorf("device %s is still not provisioned", deviceID)
}
