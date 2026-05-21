<?php

namespace App\Http\Controllers;

use App\Models\BotConfig;
use App\Models\CommandHistory;
use App\Models\User;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

class InternalApiController extends Controller
{
    /**
     * Returns per-user SMTP + AI configs for the Go bot.
     * IMAP is shared and lives in the bot's own .env.
     */
    public function botUsers(Request $request): JsonResponse
    {
        $secret = config('services.bot.secret');

        if (! $secret || $request->header('X-Bot-Secret') !== $secret) {
            abort(401);
        }

        $configs = BotConfig::with('user')
            ->where('is_active', true)
            ->get()
            ->map(fn (BotConfig $c) => [
                'user_id'           => $c->user_id,
                'user_email'        => $c->user->email,
                'smtp_host'         => $c->smtp_host,
                'smtp_port'         => $c->smtp_port,
                'smtp_username'     => $c->smtp_username,
                'smtp_password'     => $c->getSmtpPasswordDecrypted(),
                'smtp_security'     => $c->smtp_security,
                'smtp_from_name'    => $c->smtp_from_name,
                'smtp_from_address' => $c->smtp_from_address,
                'ai_provider'       => $c->ai_provider,
                'ai_api_url'        => $c->ai_api_url,
                'ai_api_key'        => $c->getAiApiKeyDecrypted(),
                'ai_model'          => $c->ai_model,
            ]);

        return response()->json($configs);
    }

    /**
     * Returns SMTP + AI config for a single sender email.
     * Called per incoming message to look up the sender's credentials.
     */
    public function userConfig(Request $request): JsonResponse
    {
        $secret = config('services.bot.secret');

        if (! $secret || $request->header('X-Bot-Secret') !== $secret) {
            abort(401);
        }

        $email = $request->query('email');
        if (! $email) {
            abort(400);
        }

        // Match by account email or smtp_from_address
        $config = BotConfig::with('user')
            ->where('is_active', true)
            ->where(fn ($q) => $q
                ->whereHas('user', fn ($u) => $u->where('email', $email))
                ->orWhere('smtp_from_address', $email)
            )
            ->first();

        if (! $config) {
            return response()->json(['found' => false]);
        }

        return response()->json([
            'found'             => true,
            'user_id'           => $config->user_id,
            'user_email'        => $config->user->email,
            'smtp_host'         => $config->smtp_host,
            'smtp_port'         => $config->smtp_port,
            'smtp_username'     => $config->smtp_username,
            'smtp_password'     => $config->getSmtpPasswordDecrypted(),
            'smtp_security'     => $config->smtp_security,
            'smtp_from_name'    => $config->smtp_from_name,
            'smtp_from_address' => $config->smtp_from_address,
            'ai_provider'       => $config->ai_provider,
            'ai_api_url'        => $config->ai_api_url,
            'ai_api_key'        => $config->getAiApiKeyDecrypted(),
            'ai_model'          => $config->ai_model,
        ]);
    }

    /**
     * Called by the Go bot after processing a command.
     * Matches the sender email to a registered user and stores the history entry.
     */
    public function storeHistory(Request $request): JsonResponse
    {
        $secret = config('services.bot.secret');

        if (! $secret || $request->header('X-Bot-Secret') !== $secret) {
            abort(401);
        }

        $data = $request->validate([
            'sender_email' => ['required', 'email'],
            'tag'          => ['required', 'string', 'max:32'],
            'payload'      => ['required', 'string'],
            'response'     => ['nullable', 'string'],
            'status'       => ['required', 'in:ok,error'],
        ]);

        // Match sender to a registered user by their smtp_from_address or account email
        $user = User::where('email', $data['sender_email'])->first()
            ?? User::whereHas('botConfig', fn ($q) => $q->where('smtp_from_address', $data['sender_email']))->first();

        if (! $user) {
            // Unknown sender – store without user association using user_id 0 would violate FK.
            // Just acknowledge silently; unregistered senders don't get history.
            return response()->json(['status' => 'skipped']);
        }

        CommandHistory::create([
            'user_id'      => $user->id,
            'tag'          => $data['tag'],
            'payload'      => $data['payload'],
            'response'     => $data['response'],
            'sender_email' => $data['sender_email'],
            'status'       => $data['status'],
        ]);

        return response()->json(['status' => 'ok']);
    }
}