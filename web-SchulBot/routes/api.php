<?php

use App\Http\Controllers\InternalApiController;
use Illuminate\Support\Facades\Route;

Route::get('/internal/bot-users', [InternalApiController::class, 'botUsers']);