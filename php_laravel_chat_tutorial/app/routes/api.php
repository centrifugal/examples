<?php

use App\Http\Controllers\CentrifugoProxyController;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Route;

/*
|--------------------------------------------------------------------------
| API Routes
|--------------------------------------------------------------------------
|
| Here is where you can register API routes for your application. These
| routes are loaded by the RouteServiceProvider within a group which
| is assigned the "api" middleware group. Enjoy building your API!
|
*/

Route::middleware('auth:sanctum')->get('/user', function (Request $request) {
    return $request->user();
});

Route::group([
    'middleware' => 'auth',
], function () {
    Route::post('/centrifugo/connect', [CentrifugoProxyController::class, 'connect']);
    Route::post('/centrifugo/publish', [CentrifugoProxyController::class, 'publish']);
    Route::post('/centrifugo/subscribe', [CentrifugoProxyController::class, 'subscribe']);
});
