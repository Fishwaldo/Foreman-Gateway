package models

import (
	"context"
	"crypto/tls"
	"errors"

	"net/http"
	"net/url"
	"time"

	"github.com/beego/beego/v2/client/cache"
	"github.com/beego/beego/v2/core/config"
	"github.com/beego/beego/v2/core/logs"
	"github.com/shurcooL/graphql"
)

var (
	hgcache cache.Cache
	nodescache cache.Cache
	graphqlclient *graphql.Client
)

func init() {
	hgcache, _ = cache.NewCache("memory", `{"interval":20}`);
	nodescache, _ = cache.NewCache("memory", `{"interval":20}`);
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	var host string;
	var user string;
	var pass string;
	var err error;
	if host, err = config.String("foreman::host"); err != nil {
		logs.Critical("Foreman host is not Set", err);
	}
	if user, err = config.String("foreman::username"); err != nil {
		logs.Critical("Foreman Username is not Set", err);
	}
	if pass, err = config.String("foreman::pat"); err != nil {
		logs.Critical("Foreman PAT is not Set", err);
	}
	var foremanurl = url.URL{
		Scheme: "https",
		User: url.UserPassword(user, pass),
		Host: host,
		Path: "/api/graphql",
	}
	logs.Informational("Using ", foremanurl.String(), "For Foreman");
	graphqlclient = graphql.NewClient(foremanurl.String(), &http.Client{Transport: tr})

	config.DefaultString("rundeck::osuser", "root");

}

type FM_Hosts struct {
	Hosts struct {
		TotalCount graphql.Int
		Edges      []struct {
			Node struct {
				Id      graphql.ID
				Name    graphql.String
				Ip      graphql.String
				Ip6     graphql.String
				Enabled graphql.Boolean
				Managed graphql.Boolean
				Domain  struct {
					Name graphql.String
				}
				Location struct {
					Name graphql.String
				}
				Hostgroup struct {
					ID   graphql.ID
					Name graphql.String
				}
				Operatingsystem struct {
					Name  graphql.String
					Major graphql.String
					Minor graphql.String
					Type  graphql.String
				}
				Architecture struct {
					Name graphql.String
				}
			}
		}
	} `graphql:"hosts"`
}

type FM_HostGroup struct {
	HostGroup struct {
		ID     graphql.ID
		Name   graphql.String
		Title  graphql.String
		Parent struct {
			ID graphql.ID
		}
	} `graphql:"hostgroup(id:$id)"`
}

type RD_Hosts struct {
	HostGroups  []string `json:"tags"`
	Location	string `json:"location"`
	OsFamily    string `json:"osFamily"`
	Username    string `json:"username"`
	OsVersion   string `json:"osVersion"`
	OsArch      string `json:"osArch"`
	Description string `json:"description"`
	Hostname    string `json:"hostname"`
	Nodename    string `json:"-"`
	OsName      string `json:"osName"`
}

func GetHosts() (fm FM_Hosts, err error) {
	if exists, _  := nodescache.IsExist(context.Background(), "NodeCache"); exists {
		tmpfm, _  := nodescache.Get(context.Background(), "NodeCache");
		return tmpfm.(FM_Hosts), nil;
	}
	var tmpfm FM_Hosts;
	logs.Info("Gettting Host List");
	err = graphqlclient.Query(context.Background(), &tmpfm, nil)
	if err != nil {
		logs.Error(err);
		return tmpfm, err;
	}
	nodescache.Put(context.Background(), "NodeCache", tmpfm, time.Minute * 5);
	return tmpfm, nil;
}

func (fm *FM_Hosts) getHostGroupsTag(id graphql.ID) (result []string, err error) {
	tags := make([]string, 0);
	for _, v := range fm.Hosts.Edges {
		if v.Node.Id != id {
			continue
		}
		hgid := v.Node.Hostgroup.ID.(string);
		for {
			var val string;
			var err error;
			if val, hgid, err = fm.getHostGroup(hgid); err != nil {
				return []string{}, errors.New("cant get HostID");
			} else {
				tags = append(tags, val);
				if (hgid == "") {
					break;
				}
			}
		}
	}
	return tags, nil;
}


func (fm *FM_Hosts) getHostGroup(hgid string) (result string, tag string, err error) {
	ret, _ := hgcache.IsExist(context.Background(), hgid)
	if ret {
		val, _ := hgcache.Get(context.Background(), hgid);
		newval, ok := val.(FM_HostGroup)
		if (ok) {
			var val string
			if (newval.HostGroup.Parent.ID != nil) {
				val = newval.HostGroup.Parent.ID.(string)
			} else {
				val = "";
			}
			return string(newval.HostGroup.Title), val, nil;
		} else {
			return "", "", errors.New("cant Cast Cached Value to a HostGroup");
		}

	} else {

		var q FM_HostGroup
		variables := map[string]interface{}{
			"id": graphql.String(hgid),
		}
		logs.Info("Getting HostGroup ID", hgid);
		err := graphqlclient.Query(context.Background(), &q, variables)
		if err != nil {
			return "", "", err;
		} else {
			hgcache.Put(context.Background(), hgid, q, time.Hour*24)
			var val string
			if (q.HostGroup.Parent.ID != nil) {
				val = q.HostGroup.Parent.ID.(string)
			} else {
				val = "";
			}
			return string(q.HostGroup.Title),val, nil;
		}
	}
}

func (fm *FM_Hosts) ToRD_Hosts() (result map[string]RD_Hosts, err error) {

	ret := make(map[string]RD_Hosts)
	osUser, err := config.String("rundeck::osuser");
	if (err != nil) {
		osUser = "root";
	}
	for _, v := range fm.Hosts.Edges {
		var ip string;
		if len(v.Node.Ip) > 0 {
			ip = string(v.Node.Ip)
		} else { 
			ip = string(v.Node.Name)
		}
		var rd = RD_Hosts{
			Nodename:    string(v.Node.Name),
			Hostname:    ip,
			OsFamily:    string(v.Node.Operatingsystem.Type),
			OsName:      string(v.Node.Operatingsystem.Name),
			OsVersion:   string(v.Node.Operatingsystem.Major),
			OsArch:      string(v.Node.Architecture.Name),
			Username:    osUser,
			Description: string(v.Node.Name),
			Location: string(v.Node.Location.Name),
		}
		if tags, err := fm.getHostGroupsTag(v.Node.Id); err == nil {
			rd.HostGroups = tags;
		}
		ret[string(v.Node.Name)] = rd;
	}
	return ret, nil
}
