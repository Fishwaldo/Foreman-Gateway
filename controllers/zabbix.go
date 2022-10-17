package controllers

import (
	"Foreman-Gateway/models"
	"fmt"
	"strconv"
	"time"
	"strings"

	"github.com/beego/beego/v2/core/config"
	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/gorilla/feeds"
)

// Operations about object
type ZabbixController struct {
	beego.Controller
}


// @Title GetAll
// @Description get all objects
// @Success 200 {object} models.Object
// @Failure 403 :objectId is empty
// @router / [get]
func (o *ZabbixController) GetAll() {

	var q []models.Problems
	var err error;
	if q, err = models.GetProblems(); err != nil {
		logs.Warn("Failed to Get All Problems");
		o.Data["content"] = "Failed to Get All Hosts";
		o.Abort("500")
	} else {
		host, _ := config.String("zabbix::host"); 
		feed := &feeds.Feed{
			Title:       "Zabbix Problems",
			Link:        &feeds.Link{Href: host},
			Description: "Zabbix Problems",
			Created: time.Now(),
		}
		for _, item := range q {
			ts, _ := strconv.ParseInt(item.LastChange, 10, 0)
			feed.Items = append(feed.Items, &feeds.Item{
				Title:       fmt.Sprintf("%s: %s", item.Hosts[0].Name, item.Description),
				Link:        &feeds.Link{Href: fmt.Sprintf("%s/tr_events.php?triggerid=%s&eventid=%s", host, item.TriggerID, item.LastEvent.EventID)},
				Description: item.Description,
				Created:     time.Unix(ts, 0),
			})
		}
		rss, err := feed.ToRss()
		if err != nil {
			logs.Warn("Failed to Generate RSS");
			o.Data["content"] = "Failed to Generate RSS";
			o.Abort("500")
		}
		fmt.Printf("%+v", rss)
		o.Ctx.ResponseWriter.Header().Set("Content-Type", "application/xml")
		o.Ctx.WriteString(strings.TrimPrefix(rss, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>"))
//		o.Resp(rss)

	}
}


// @Title Update
// @Description update the object
// @Param	objectId		path 	string	true		"The objectid you want to update"
// @Param	body		body 	models.Object	true		"The body"
// @Success 200 {object} models.Object
// @Failure 403 :objectId is empty
// @router /:objectId [put]
//func (o *ObjectController) Put() {
//	objectId := o.Ctx.Input.Param(":objectId")
//	var ob models.Object
//	json.Unmarshal(o.Ctx.Input.RequestBody, &ob)
//
//	err := models.Update(objectId, ob.Score)
//	if err != nil {
//		o.Data["json"] = err.Error()
//	} else {
//		o.Data["json"] = "update success!"
//	}
//	o.ServeJSON()
//}

// @Title Delete
// @Description delete the object
// @Param	objectId		path 	string	true		"The objectId you want to delete"
// @Success 200 {string} delete success!
// @Failure 403 objectId is empty
// @router /:objectId [delete]
//func (o *ObjectController) Delete() {
//	objectId := o.Ctx.Input.Param(":objectId")
//	models.Delete(objectId)
//	o.Data["json"] = "delete success!"
//	o.ServeJSON()
//}
