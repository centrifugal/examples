<?php

namespace App\Http\Controllers;

use App\Models\Message;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Http\Response;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\Log;
use stdClass;
use Throwable;

class CentrifugoProxyController extends Controller
{
    public function connect()
    {
        return new JsonResponse([
            'result' => [
                'user' => (string) Auth::user()->id
            ]
        ]);
    }

    public function publish(Request $request)
    {
        $requestData = $request->json()->all();
        $status = Response::HTTP_OK;

        try {
            Message::create([
                'sender_id' => $requestData["user"],
                'message' => $requestData["data"]["message"],
                'room_id' => $requestData["data"]["room_id"],
            ]);
        } catch (Throwable $e) {
            Log::error($e->getMessage());
            $status = Response::HTTP_INTERNAL_SERVER_ERROR;
        }

        return new JsonResponse([
            'result' => new stdClass()
        ], $status);
    }

    public function subscribe()
    {
        return new JsonResponse([
            'result' => new stdClass()
        ]);
    }
}
