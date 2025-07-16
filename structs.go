package main

import (
	"encoding/xml"
	"time"
)

type Response struct {
	Servers []*ServerInfo `json:"servers"`
	Errors  []string      `json:"errors"`
}

type WorkerResult struct {
	URL    string
	Status *Status
	Error  error
}

type ServerInfo struct {
	Name       string `json:"name"`
	CommonName string `json:"commonName"`
	Status     bool   `json:"status"`
	Order      int    `json:"order"`
}

type Status struct {
	XMLName                 xml.Name `xml:"Status"`
	LoginTierLastNumbers    string   `xml:"logintierlastnumbers"`
	LoginTiers              string   `xml:"logintiers"`
	QueueNames              string   `xml:"queuenames"`
	AllowBillingRole        string   `xml:"allow_billing_role"`
	QueueURLs               string   `xml:"queueurls"`
	LastAssignedQueueNumber string   `xml:"lastassignedqueuenumber"`
	Name                    string   `xml:"name"`
	FarmID                  string   `xml:"farmid"`
	DenyAdminRole           string   `xml:"deny_admin_role"`
	WorldFull               string   `xml:"world_full"`
	WaitHint                string   `xml:"wait_hint"`
	WePermaDeath            string   `xml:"we_perma_death"`
	AllowAdminRole          string   `xml:"allow_admin_role"`
	NowServingQueueNumber   string   `xml:"nowservingqueuenumber"`
	DenyBillingRole         string   `xml:"deny_billing_role"`
	LoginTierMultipliers    string   `xml:"logintiermultipliers"`
	LoginServers            string   `xml:"loginservers"`
	WorldPVPPermission      string   `xml:"world_pvppermission"`
}

type ArrayOfDatacenterStruct struct {
	XMLName           xml.Name           `xml:"ArrayOfDatacenterStruct"`
	DatacenterStructs []DatacenterStruct `xml:"DatacenterStruct"`
}

type DatacenterStruct struct {
	KeyName    string         `xml:"KeyName"`
	Datacenter DatacenterWrap `xml:"Datacenter"`
}

type DatacenterWrap struct {
	CachedAt   time.Time      `xml:"cachedAt"`
	Datacenter DatacenterInfo `xml:"datacenter>Datacenter"`
}

type DatacenterInfo struct {
	Name                        string  `xml:"Name"`
	Worlds                      []World `xml:"Worlds>World"`
	AuthServer                  string  `xml:"AuthServer"`
	PatchServer                 string  `xml:"PatchServer"`
	LauncherConfigurationServer string  `xml:"LauncherConfigurationServer"`
}

type World struct {
	Name            string `xml:"Name"`
	LoginServerUrl  string `xml:"LoginServerUrl"`
	ChatServerUrl   string `xml:"ChatServerUrl"`
	StatusServerUrl string `xml:"StatusServerUrl"`
	Language        string `xml:"Language,omitempty"`
	Order           int    `xml:"Order"`
}
