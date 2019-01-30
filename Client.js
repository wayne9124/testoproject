var mqtt = require('mqtt')
var jwt = require('jsonwebtoken')
var express = require('express');
var app = express();


var tokenClinet=null;
var config = require('./config')
var client  = mqtt.connect('mqtt://127.0.0.1:1883')
var dve=({
    "id": null,
    "name": "three",
    "password": "147258"
})

app.set('Req', config.secret_Req)
app.set('Reg', config.secret_Reg)

client.on('connect', function () {
    var string=JSON.stringify(dve);
    var locatname = JSON.parse(string)
    client.subscribe('Reg-client', function (err) {});///接收token
    client.subscribe(locatname.name+'ErrorReport', function (err) {})
    var Regpack = jwt.sign(dve, app.get('Reg'), {}) ///打包註冊資料
    client.publish('Reg-sever', Regpack.toString())///推送註冊資料(加密過的) 
    
})

client.on('message', function (topic, message){///收到回應
    if(topic == 'Reg-client'){
        tokenClinet=message.toString()
        console.log(tokenClinet)
        var Newuser= ({
            "id": null,
            "name": "bird",
            "password": "456789",
            "token" : tokenClinet
        })
        var paloy = jwt.sign(Newuser, app.get('Req'), {}) ///打包註冊資料
        client.publish('Req-sever', paloy.toString()) 
    } else{
        console.log(message.toString())
    }
})
