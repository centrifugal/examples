import jwt
from django.conf import settings


def get_connection_token(user, info=''):
    user_pk = str(user.pk) if user.is_authenticated else ""
    return jwt.encode({"user": user_pk}, settings.CENTRIFUGE_SECRET).decode()


def main(request):
    return dict(
        CENTRIFUGE_SOCKJS_ENDPOINT=settings.CENTRIFUGE_ADDRESS + '/connection/sockjs',
        CENTRIFUGE_WS_ENDPOINT=settings.CENTRIFUGE_ADDRESS + '/connection/websocket',
        CENTRIFUGE_TOKEN=get_connection_token(request.user),
    )
