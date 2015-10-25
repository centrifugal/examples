Centrifuge with Django
======================

Simple demo site to display events on Google map in real-time.

To run demo

1) Clone this repo

2) Install requirements

```bash
pip install -r requirements.txt
```

3) Add your Centrifuge (must be already running) parameters in `settings.py`:

```python
CENTRIFUGE_ADDRESS = 'http://localhost:8000'
CENTRIFUGE_SECRET = 'secret'
CENTRIFUGE_TIMEOUT = 5
```

4) Use this config in Centrifugo:

```
{
  "secret": "secret",
  "namespaces": [
    {
      "name": "public",
      "anonymous": true
    }
  ]
}
```

Make sure that `anonymous` access allowed in project settings in Centrifuge - as all Django users anonymous in our case.

5) Run Django server

```bash
python manage.py runserver 0:8080
```

6) Go to http://localhost:8080


You will see a map and you can start sending events into `map` channel:

```bash
python manage.py publish --lat=34 --long=54 --content="test"
```

Where:

`--lat` - latitude
`--long` - longitude
`--content` - content of Info Window

Or via `cent` console client:

```bash
echo '{"channel": "public:map", "data": {"lat": 33, "long": 55, "content": "I am testing Centrifuge"}}'|cent map publish
```

The contents of my `~/.centrc` file in this case:

```bash
[map]
address = http://localhost:8000/api
secret = secret
```

After this all connected clients will see new event on map.
