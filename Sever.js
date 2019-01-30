var mqtt = require('mqtt')
var bodyParser = require('body-parser')
var jwt = require('jsonwebtoken')
var express = require('express');
var app = express();
var mysql = require('mysql');



var config = require('./config')
var client  = mqtt.connect('mqtt://127.0.0.1:1883')
var Usermanger = mysql.createConnection({
    host: 'localhost',
    user: 'UserManger',
    password: '123456',
    database: 'users'
  });
var DevManger = mysql.createConnection({
    host: 'localhost',
    user: 'DevManger',
    password: '123456',
    database: 'dev'
  });
app.set('Req', config.secret_Req)
app.set('Reg', config.secret_Reg)


client.on('connect', function () {
  client.subscribe('update', function (err) {});
  client.subscribe('Reg-sever', function (err) {});
  client.subscribe('Req-sever', function (err) {});
})
 
client.on('message', function (topic, message) {
  if(topic == 'Reg-sever'){
    console.log(message.toString())
    console.log(topic.toString())
    if (message) {//有訊息的話，確認設備身分
      jwt.verify(message.toString(), app.get('Reg'), function (err, decoded1) {
        if (err) {
          console.log({success: false, message: 'Failed to authenticate token.'})
        } else {
         console.log("success: true")
         console.log(decoded1)
         console.log(decoded1.name)
         console.log(decoded1.password)
         ///DB驗證開始
         ///DB驗證完成
         var token = jwt.sign(decoded1, app.get('Req'), {})///用req來加密
         client.publish('Reg-client', token.toString())
        }
      })
     } else {
        console.log({
        success: false,
        message: 'token error.'})}
  }else if(topic == 'Req-sever'){
    jwt.verify(message.toString(), app.get('Req'), function (err, decoded2){
      if (err) {
        console.log({success: false, message: 'Failed to authenticate token.'})
      } else {
       console.log("success: true")
       console.log(decoded2)
       console.log(decoded2.name)
       console.log(decoded2.password)
       console.log(decoded2.token)
       jwt.verify(decoded2.token, app.get('Req'), function (err, decoded){
        if (err) {
          console.log({success: false, message: 'Failed to authenticate token.'})
        } else {
          console.log("DVE verify success: true")
          console.log("DB ACTIVE star")
        }
       })
      }

    })
  }
})
