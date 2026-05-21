<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;
use Inertia\Inertia;
use Inertia\Response;

class CommandHistoryController extends Controller
{
    public function index(Request $request): Response
    {
        $history = $request->user()
            ->commandHistory()
            ->orderByDesc('created_at')
            ->paginate(20)
            ->through(fn ($item) => [
                'id'           => $item->id,
                'tag'          => $item->tag,
                'payload'      => $item->payload,
                'response'     => $item->response,
                'sender_email' => $item->sender_email,
                'status'       => $item->status,
                'created_at'   => $item->created_at->toIso8601String(),
            ]);

        return Inertia::render('bot/History', [
            'history' => $history,
        ]);
    }
}