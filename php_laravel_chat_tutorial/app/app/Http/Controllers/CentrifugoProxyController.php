<?php

namespace App\Http\Controllers;

use Illuminate\Http\JsonResponse;
use Illuminate\Support\Facades\Auth;

class CentrifugoProxyController extends Controller
{
    public function connect()
    {
        return new JsonResponse([
            'result' => [
                'user' => (string) Auth::user()->id,
                'channels' => ["personal:#".Auth::user()->id],
            ]
        ]);
    }
}
