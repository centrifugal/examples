const express = require('express');
const cookieParser = require("cookie-parser");
const sessions = require('express-session');
const morgan = require('morgan');
const axios = require('axios');

const app = express();
const port = 3000;
app.use(express.json());

const oneDay = 1000 * 60 * 60 * 24;

app.use(sessions({
  secret: "this_is_my_secret_key",
  saveUninitialized: true,
  cookie: { maxAge: oneDay },
  resave: false
}));
app.use(cookieParser());
app.use(express.urlencoded({ extended: true }))
app.use(express.json())
app.use(express.static('static'));
app.use(morgan('dev'));

app.post('/centrifugo/connect', (req, res) => {
  console.log(req.body);
  console.log(req.cookies);

  if (req.session.userid) {
    res.json({
      "result": {
        "user": req.session.userid
      }
    });
  } else
    res.json({
      "disconnect": {
        "code": 1000,
        "reason": "unauthorized",
        "reconnect": false
      }
    });
});

const myusername = 'demo-user'
const mypassword = 'demo-pass'

app.post('/login', (req, res) => {
  if (req.body.username == myusername && req.body.password == mypassword) {
    req.session.userid = req.body.username;
    res.redirect('/');
  } else {
    res.send('Invalid username or password');
  }
});

app.get('/logout', (req, res) => {
  req.session.destroy();
  res.redirect('/');
});

app.get('/', (req, res) => {
  if (req.session.userid) {
    res.sendFile('views/app.html', { root: __dirname });
  } else
    res.sendFile('views/login.html', { root: __dirname })
});

const centrifugoApiClient = axios.create({
  baseURL: `http://centrifugo:8000/api`,
  headers: {
    Authorization: `apikey my_api_key`,
    'Content-Type': 'application/json',
  },
});

setInterval(async () => {
  try {
    await centrifugoApiClient.post('', {
      method: 'publish',
      params: {
        channel: '#' + myusername, // construct personal channel name.
        data: {
          time: Math.floor(new Date().getTime() / 1000),
        },
      },
    });
  } catch (e) {
    console.error(e.message);
  }
}, 5000);

app.listen(port, () => {
  console.log(`Example app listening at http://localhost:${port}`);
});
