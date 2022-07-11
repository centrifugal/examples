from django.shortcuts import render
from django.http import JsonResponse
from django.views.decorators.csrf import csrf_exempt

import logging

# Get an instance of a logger
logger = logging.getLogger(__name__)


def index(request):
    return render(request, 'chat/index.html')


def room(request, room_name):
    return render(request, 'chat/room.html', {
        'room_name': room_name
    })


@csrf_exempt
def connect(request):
    logger.debug(request.body)
    response = {
        'result': {
            'user': 'tutorial-user'
        }
    }
    return JsonResponse(response)


@csrf_exempt
def publish(request):
    logger.debug(request.body)
    response = {
        'result': {}
    }
    return JsonResponse(response)


@csrf_exempt
def subscribe(request):
    logger.debug(request.body)
    response = {
        'result': {}
    }
    return JsonResponse(response)
