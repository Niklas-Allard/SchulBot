<?php

namespace App\Http\Controllers;

use App\Models\CommandHistory;
use App\Models\User;
use Illuminate\Http\RedirectResponse;
use Illuminate\Http\Request;
use Inertia\Inertia;
use Inertia\Response;

class AdminController extends Controller
{
    public function users(): Response
    {
        $users = User::with('botConfig')
            ->orderByDesc('created_at')
            ->get()
            ->map(fn ($user) => [
                'id'         => $user->id,
                'name'       => $user->name,
                'email'      => $user->email,
                'role'       => $user->role,
                'bot_active' => $user->botConfig?->is_active ?? false,
                'bot_set_up' => $user->botConfig !== null,
                'created_at' => $user->created_at->toIso8601String(),
            ]);

        $stats = [
            'total_users'    => User::count(),
            'active_bots'    => \App\Models\BotConfig::where('is_active', true)->count(),
            'commands_today' => CommandHistory::whereDate('created_at', today())->count(),
            'commands_total' => CommandHistory::count(),
        ];

        return Inertia::render('admin/Users', [
            'users' => $users,
            'stats' => $stats,
        ]);
    }

    public function toggleBot(Request $request, User $user): RedirectResponse
    {
        $config = $user->botConfig;
        if ($config) {
            $config->update(['is_active' => ! $config->is_active]);
        }

        return back();
    }

    public function makeAdmin(User $user): RedirectResponse
    {
        $user->update(['role' => 'admin']);

        return back();
    }

    public function destroyUser(User $user): RedirectResponse
    {
        $user->delete();

        return back();
    }
}