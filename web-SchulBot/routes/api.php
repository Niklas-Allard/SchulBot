<?php

use App\Http\Controllers\InternalApiController;
use Illuminate\Support\Facades\Route;

Route::get('/internal/bot-users', [InternalApiController::class, 'botUsers']);
Route::get('/internal/user-config', [InternalApiController::class, 'userConfig']);
Route::post('/internal/history', [InternalApiController::class, 'storeHistory']);