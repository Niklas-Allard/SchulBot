<?php

use App\Http\Controllers\AdminController;
use App\Http\Controllers\BotConfigController;
use App\Http\Controllers\CommandHistoryController;
use App\Http\Controllers\DashboardController;
use App\Http\Middleware\AdminMiddleware;
use Illuminate\Support\Facades\Route;

Route::inertia('/', 'Welcome')->name('home');

Route::middleware(['auth', 'verified'])->group(function () {
    Route::inertia('/guide', 'Guide')->name('guide');
    Route::get('dashboard', [DashboardController::class, 'index'])->name('dashboard');

    Route::get('bot/setup', [BotConfigController::class, 'edit'])->name('bot.setup');
    Route::post('bot/setup', [BotConfigController::class, 'update'])->name('bot.setup.update');

    Route::get('bot/history', [CommandHistoryController::class, 'index'])->name('bot.history');
});

Route::middleware(['auth', 'verified', AdminMiddleware::class])->prefix('admin')->name('admin.')->group(function () {
    Route::get('users', [AdminController::class, 'users'])->name('users');
    Route::post('users/{user}/toggle-bot', [AdminController::class, 'toggleBot'])->name('users.toggle-bot');
    Route::post('users/{user}/make-admin', [AdminController::class, 'makeAdmin'])->name('users.make-admin');
    Route::delete('users/{user}', [AdminController::class, 'destroyUser'])->name('users.destroy');
});

require __DIR__.'/settings.php';
