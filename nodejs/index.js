/* npm install express path jscent */

var express = require("express");
var Client = require("jscent");
var path = require('path');

var host = "0.0.0.0";
var port = 8080;

var app = express();
app.use(express.static(__dirname + "/static"));
app.set('views', path.join(__dirname, 'views'));
app.set('view engine', 'ejs');

var timestamp = parseInt(new Date().getTime()/1000).toString();
var user = "42";
var Token = new Client.Token("secret_key");
Token = Token.clientToken(user, timestamp, "");

app.get('/', function(req, res){ 
	res.render('index',{Token: Token, Timestamp: timestamp, User: user});
});

app.listen(port, host);
console.log("Server ready at "+host+" "+port)
