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
    console.log('---------------Incoming DevReg information------------------')
    console.log(message.toString())
    console.log(topic.toString())
    if (message) {//有訊息的話，確認設備身分
      jwt.verify(message.toString(), app.get('Reg'), function (err, DevReg) {
        if (err) {
          console.log({success: false, message: '-----------------------Failed to authenticate token.-----------------\n\n.'})
          } else {//查
          console.log("Dev name : "+DevReg.name)
          console.log("Dev password : "+DevReg.password)
          console.log("------------------------------------------------------------.")

          var  sql = 'select name, password from dev';

          DevManger.query(sql,function (err, result) {///資料庫查詢
           if(err){
            console.log('[SELECT ERROR] - ',err.message);
            return;
           }
           var string=JSON.stringify(result); 
           var data = JSON.parse(string) // iterate over each element in the array
           var pass= false;
           for (var i = 0; i < data.length; i++){ // 資料庫查詢結果比對
              if (data[i].name == DevReg.name && data[i].password == DevReg.password){
                // we found it
                // obj[i].name is the matched result
                
                 var token = jwt.sign(DevReg, app.get('Req'), {})///用req來加密
                 client.publish('Reg-client', token.toString())
                 console.log("DVE verify success")
                 pass = true;
                  
                }
              }
             if(pass==false){
              console.log('--------------------------Error-----------------------------')
              console.log("User or PassWord wrong");
              client.publish(DevReg.name+'ErrorReport', "Dev "+DevReg.name+" : User or PassWord wrong")
              console.log('------------------------------------------------------------\n\n');  

             }

           });
          } 
        })
      }else{
        console.log({
        success: false,
        message: 'token error.'})
      }
    
    }else if(topic == 'Req-sever'){
     jwt.verify(message.toString(), app.get('Req'), function (err, NewUserReq){
      if (err) {
        console.log({success: false, message: 'Failed to authenticate token.'})
      } else {
        console.log("----------------verify success-------------------------.")
       console.log("New user name : "+NewUserReq.name)
       console.log("New user password : "+NewUserReq.password)
       console.log("req token : "+NewUserReq.token)
       jwt.verify(NewUserReq.token, app.get('Req'), function (err, decoded){
        if (err) {
          console.log({success: false, message: '----------------no-Reg Dev--------------------'})
        } else {
          
          var  addSql = 'INSERT INTO user(id,name,password) VALUES(0,?,?)';
          var  addSqlParams = [NewUserReq.name, NewUserReq.password];
          
          Usermanger.query(addSql,addSqlParams,function (err, result) {
           if(err){
           console.log('[INSERT ERROR] - ',err.message);
             return;
           }else{       
              console.log('--------------------------INSERT----------------------------');
              console.log('INSERT ID:',result.insertId);        
              console.log('INSERT ID:',result);        
              console.log('--------------------------Finish-----------------------------\n\n');  
              ///message is Buffer    
              
              
              
               }
            })
          
        }
       })
      }

    })
  }
})
