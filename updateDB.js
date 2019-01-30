var mqtt = require('mqtt')
var jwt = require('jsonwebtoken')
var express = require('express');
var app = express();


var tokenClinet=null;
var config = require('./config')
var client  = mqtt.connect('mqtt://127.0.0.1:1883')
var dve=({
    "id": null,
    "name": "Update",
    "password": "456789"
})
var string=JSON.stringify(dve);
var locatname = JSON.parse(string)
app.set('Req', config.secret_Req)
app.set('Reg', config.secret_Reg)

client.on('connect', function () {
    client.subscribe(locatname.name+'token', function (err) {});///接收token
    client.subscribe(locatname.name+'ErrorReport', function (err) {})
    var Regpack = jwt.sign(dve, app.get('Reg'), {}) ///打包註冊資料
    client.publish('Reg-sever', Regpack.toString())///推送註冊資料(加密過的) 2
    
})

client.on('message', function (topic, message){///收到回應
  if(topic == (locatname.name+'token')){
        tokenClinet=message.toString()
        console.log(tokenClinet)
        var updateuser= ({
            "id": 10,
            "name": "girl",
            "password": "115522",
            "token" : tokenClinet
        })
        var paloy = jwt.sign(updateuser, app.get('Req'), {}) ///打包註冊資料
        client.publish('Req-update', paloy.toString()) 
    } else{
        console.log(message.toString())
    }
})
