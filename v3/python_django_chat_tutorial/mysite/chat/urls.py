from django.urls import path, re_path

from . import views

urlpatterns = [
    path('', views.index, name='index'),
    re_path('room/(?P<room_name>[A-z0-9_-]+)/', views.room, name='room'),
    path('centrifugo/connect/', views.connect, name='connect'),
    path('centrifugo/subscribe/', views.subscribe, name='subscribe'),
    path('centrifugo/publish/', views.publish, name='publish'),
]
