package controllers

import (
	"github.com/astaxie/beego"
)

type baseController struct {
	beego.Controller
}

func (this *baseController) Prepare() {
	//签名验证等
	//this.Data["json"] = map[string]interface{}{"msg": "verify not pass"}
	//this.ServeJSON()
	//this.StopRun()
}
