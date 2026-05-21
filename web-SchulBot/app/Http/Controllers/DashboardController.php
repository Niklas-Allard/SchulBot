<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;
use Inertia\Inertia;
use Inertia\Response;

class DashboardController extends Controller
{
    public function index(Request $request): Response
    {
        $user = $request->user();
        $config = $user->botConfig;

        $recent = $user->commandHistory()
            ->orderByDesc('created_at')
            ->limit(5)
            ->get()
            ->map(fn ($item) => [
                'id'        => $item->id,
                'tag'       => $item->tag,
                'payload'   => $item->payload,
                'status'    => $item->status,
                'created_at' => $item->created_at->toIso8601String(),
            ]);

        return Inertia::render('Dashboard', [
            'botConfigured'  => $config !== null,
            'botActive'      => $config?->is_active ?? false,
            'commandsTotal'  => $user->commandHistory()->count(),
            'commandsToday'  => $user->commandHistory()->whereDate('created_at', today())->count(),
            'recent'         => $recent,
        ]);
    }
}