package models

import (
	"context"

	"net/http"
	"time"

	"github.com/beego/beego/v2/client/cache"
	"github.com/beego/beego/v2/core/config"
	"github.com/beego/beego/v2/core/logs"
	"github.com/cavaliercoder/go-zabbix"
)

var (
	zbcache cache.Cache
	session *zabbix.Session
)

func init() {
	zbcache, _ = cache.NewCache("memory", `{"interval":20}`);
	var host string;
	var user string;
	var pass string;
	var err error;
	if host, err = config.String("zabbix::host"); err != nil {
		logs.Critical("Zabbix host is not Set", err);
	}
	if user, err = config.String("zabbix::username"); err != nil {
		logs.Critical("Zabbix Username is not Set", err);
	}
	if pass, err = config.String("zabbix::pat"); err != nil {
		logs.Critical("Zabbix PAT is not Set", err);
	}
	logs.Informational("Using ", host, "For Zabbix");

	client := &http.Client{
		Transport: &http.Transport{},
	}

	cache := zabbix.NewSessionFileCache().SetFilePath("./zabbix_session")
	session, err = zabbix.CreateClient(host + "/api_jsonrpc.php").
		WithCache(cache).
		WithHTTPClient(client).
		WithCredentials(user, pass).
		Connect()
	if err != nil {
		logs.Critical("Zabbix Connection Failed", err);
	}

}
type Problems struct {
	Comments string `json:"comments,omitempty"`
	CorrelationMode string `json:"correlation_mode,omitempty"`
	CorrelationTag string `json:"correlation_tag,omitempty"`
	Description string `json:"description,omitempty"`
	Error string `json:"error,omitempty"`
	EventName string `json:"event_name,omitempty"`
	Expression string `json:"expression,omitempty"`
	Flags string `json:"flags,omitempty"`
	Hosts []struct {
		AutoCompress string `json:"auto_compress,omitempty"`
		CustomInterfaces string `json:"custom_interfaces,omitempty"`
		Description string `json:"description,omitempty"`
		Flags string `json:"flags,omitempty"`
		Host string `json:"host,omitempty"`
		HostID string `json:"hostid,omitempty"`
		InventoryMode string `json:"inventory_mode,omitempty"`
		IPMIAuthType string `json:"ipmi_authtype,omitempty"`
		IPMIPassword string `json:"ipmi_password,omitempty"`
		IPMIPrivilege string `json:"ipmi_privilege,omitempty"`
		IPMIUsername string `json:"ipmi_username,omitempty"`
		MainenanceFrom string `json:"maintenance_from,omitempty"`
		MaintenanceStatus string `json:"maintenance_status,omitempty"`
		MaintenanceType string `json:"maintenance_type,omitempty"`
		MaintenanceID string `json:"maintenanceid,omitempty"`
		Name string `json:"name,omitempty"`
		ProxyAddress string `json:"proxy_address,omitempty"`
		ProxyHostID string `json:"proxy_hostid,omitempty"`
		Status string `json:"status,omitempty"`
		TemplateID string `json:"templateid,omitempty"`
		TLSAccept string `json:"tls_accept,omitempty"`
		TLSConnect string `json:"tls_connect,omitempty"`
		TLSIssuer string `json:"tls_issuer,omitempty"`
		TLSSubject string `json:"tls_subject,omitempty"`
		UUID string `json:"uuid,omitempty"`
	}
	LastEvent struct {
		Acknowledged string `json:"acknowledged,omitempty"`
		Clock string `json:"clock,omitempty"`
		EventID string `json:"eventid,omitempty"`
		Name string `json:"name,omitempty"`
		NS string `json:"ns,omitempty"`
		Object string `json:"object,omitempty"`
		ObjectID string `json:"objectid,omitempty"`
		Severity string `json:"severity,omitempty"`
		Source string `json:"source,omitempty"`
		Value string `json:"value,omitempty"`
	}
	LastChange string `json:"lastchange,omitempty"`
	ManualClose string `json:"manual_close,omitempty"`
	OpData string `json:"opdata,omitempty"`
	Priority string `json:"priority,omitempty"`
	RecoveryExpression string `json:"recovery_expression,omitempty"`
	RecoveryMode string `json:"recovery_mode,omitempty"`
	State string `json:"state,omitempty"`
	Status string `json:"status,omitempty"`
	TemplateID string `json:"templateid,omitempty"`
	TriggerID string `json:"triggerid,omitempty"`
	Type string `json:"type,omitempty"`
	URL string `json:"url,omitempty"`
	UUID string `json:"uuid,omitempty"`
	Value string `json:"value,omitempty"`

}


func GetProblems() (pb []Problems, err error) {
	if exists, _  := zbcache.IsExist(context.Background(), "ProblemCache"); exists {
		tmpfm, _  := zbcache.Get(context.Background(), "ProblemCache");
		return tmpfm.([]Problems), nil;
	}
	logs.Info("Gettting Problem List");
	tp := zabbix.TriggerGetParams{
		ExpandDescription: true,
		ExpandComment: true,
		ExpandExpression: true,
		SelectHosts: zabbix.SelectExtendedOutput,
		SelectLastEvent: zabbix.SelectExtendedOutput,
		ActiveOnly: true,
		MonitoredOnly: true,
	}
	tp.RecentProblemOnly = true
	var problem []Problems
	session.Get("trigger.Get", tp, &problem)


	zbcache.Put(context.Background(), "ProblemCache", problem, time.Minute * 5);
	return problem, nil;
}

