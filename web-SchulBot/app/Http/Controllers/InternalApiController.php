<?php

namespace App\Http\Controllers;

use App\Models\BotConfig;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

class InternalApiController extends Controller
{
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
                'imap_host'         => $c->imap_host,
                'imap_port'         => $c->imap_port,
                'imap_username'     => $c->imap_username,
                'imap_password'     => $c->getImapPasswordDecrypted(),
                'imap_security'     => $c->imap_security,
                'imap_mailbox'      => $c->imap_mailbox,
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
}