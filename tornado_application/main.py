# -*- coding: utf-8 -*-
from __future__ import print_function
import time
import json
import logging

import tornado.ioloop
import tornado.web
from tornado.options import options, define
import jwt


logging.getLogger().setLevel(logging.DEBUG)


define(
    "port", default=3000, help="app port", type=int
)
define(
    "centrifuge", default='localhost:8000',
    help="centrifuge address without url scheme", type=str
)
define(
    "secret", default='', help="secret key", type=str
)


# your application's user ID
USER_ID = '2694'

# your application's connection info (optional)
INFO = {
    'first_name': 'Alexander',
    'last_name': 'Emelin'
}


class IndexHandler(tornado.web.RequestHandler):

    def get(self):
        self.render('index.html')


def get_connection_token():
    token = jwt.encode({"sub": USER_ID, "info": INFO, "exp": int(time.time()) + 10}, key=options.secret)
    return token.decode()


class SockjsHandler(tornado.web.RequestHandler):

    def get(self):
        """
        Render template with data required to connect to Centrifuge using SockJS.
        """
        self.render(
            "index_sockjs.html",
            auth_data={
                'token': get_connection_token()
            },
            centrifuge_address=options.centrifuge
        )


class WebsocketHandler(tornado.web.RequestHandler):

    def get(self):
        """
        Render template with data required to connect to Centrifuge using Websockets.
        """
        self.render(
            "index_websocket.html",
            auth_data={
                'token': get_connection_token()
            },
            centrifuge_address=options.centrifuge
        )


class CentrifugeSubscribeHandler(tornado.web.RequestHandler):
    """
    Allow all users to subscribe on channels they want.
    """
    def check_xsrf_cookie(self):
        pass

    def post(self):
        try:
            data = json.loads(self.request.body)
        except ValueError:
            raise tornado.web.HTTPError(403)

        client = data.get("client", "")
        channels = data.get("channels", [])

        logging.info("{0} wants to subscribe on {1}".format(client, ", ".join(channels)))

        channel_data = []

        for channel in channels:
                channel_data.append({
                    "channel": channel,
                    "token": jwt.encode({
                        "client": client,
                        "channel": channel,
                        "info": {
                            'extra': 'extra for ' + channel
                        }, 
                        "exp": int(time.time()) + 10
                    }, key=options.secret).decode()
                })

        # but here we allow to join any private channel and return additional
        # JSON info specific for channel
        self.set_header('Content-Type', 'application/json; charset="utf-8"')
        self.write(json.dumps({
            "channels": channel_data
        }))


class CentrifugeRefreshHandler(tornado.web.RequestHandler):
    """
    Allow all users to subscribe on channels they want.
    """
    def check_xsrf_cookie(self):
        pass

    def post(self):
        #raise tornado.web.HTTPError(403)
        logging.info("client wants to refresh its connection parameters")
        self.set_header('Content-Type', 'application/json; charset="utf-8"')
        self.write(json.dumps({
            'token': get_connection_token()
        }))



# Connect proxy example handler.
class CentrifugoConnectHandler(tornado.web.RequestHandler):

    def check_xsrf_cookie(self):
        pass

    def post(self):
        logging.info(self.request.body)
        self.set_header('Content-Type', 'application/json; charset="utf-8"')
        data = json.dumps({
            'result': {
                'user': '56',
                'expire_at': int(time.time()) + 10
            }
            # 'error': {
            #     'code': 1000,
            #     'message': 'custom error'
            # },
            # 'disconnect': {
            #     'code': 4000,
            #     'reconnect': False,
            #     'reason': 'custom disconnect'
            # }
        })
        logging.info(data)
        self.write(data)

# Refresh proxy example handler.
class CentrifugoRefreshHandler(tornado.web.RequestHandler):

    def check_xsrf_cookie(self):
        pass

    def post(self):
        logging.info(self.request.body)
        self.set_header('Content-Type', 'application/json; charset="utf-8"')
        data = json.dumps({
            'result': {
                'expire_at': int(time.time()) + 10
            }
        })
        logging.info(data)
        self.write(data)


# RPC proxy example handler.
class CentrifugoRPCHandler(tornado.web.RequestHandler):

    def check_xsrf_cookie(self):
        pass

    def post(self):
        logging.info(self.request.body)
        self.set_header('Content-Type', 'application/json; charset="utf-8"')
        try:
            rpcRequest = json.loads(self.request.body)
        except ValueError:
            raise tornado.web.HTTPError(403)
        data = json.dumps({
            'result': {
                'data': {"answer": 2019}
            }
        })
        logging.info(data)
        self.write(data)


# Subscribe proxy example handler.
class CentrifugoSubscribeHandler(tornado.web.RequestHandler):

    def check_xsrf_cookie(self):
        pass

    def post(self):
        logging.info(self.request.body)
        self.set_header('Content-Type', 'application/json; charset="utf-8"')
        try:
            subscribeRequest = json.loads(self.request.body)
        except ValueError:
            raise tornado.web.HTTPError(403)
        data = json.dumps({
            'result': {}
        })
        logging.info(data)
        self.write(data)


# Publish proxy example handler.
class CentrifugoPublishHandler(tornado.web.RequestHandler):

    def check_xsrf_cookie(self):
        pass

    def post(self):
        logging.info(self.request.body)
        self.set_header('Content-Type', 'application/json; charset="utf-8"')
        try:
            publishRequest = json.loads(self.request.body)
        except ValueError:
            raise tornado.web.HTTPError(403)
        data = json.dumps({
            'result': {}
        })
        logging.info(data)
        self.write(data)


def run():
    options.parse_command_line()
    app = tornado.web.Application(
        [
            (r'/', IndexHandler),
            (r'/sockjs', SockjsHandler),
            (r'/ws', WebsocketHandler),
            (r'/centrifuge/subscribe', CentrifugeSubscribeHandler),
            (r'/centrifuge/refresh', CentrifugeRefreshHandler),
            (r'/centrifugo/connect', CentrifugoConnectHandler),
            (r'/centrifugo/refresh', CentrifugoRefreshHandler),
            (r'/centrifugo/rpc', CentrifugoRPCHandler),
            (r'/centrifugo/subscribe', CentrifugoSubscribeHandler),
            (r'/centrifugo/publish', CentrifugoPublishHandler),
        ],
        debug=True
    )
    app.listen(options.port)
    logging.info("app started, visit http://localhost:%s" % options.port)
    tornado.ioloop.IOLoop.instance().start()


def main():
    try:
        run()
    except KeyboardInterrupt:
        pass


if __name__ == '__main__':
    main()
