<?php

namespace App\Http\Controllers;

use App\Models\BotConfig;
use Illuminate\Http\RedirectResponse;
use Illuminate\Http\Request;
use Inertia\Inertia;
use Inertia\Response;

class BotConfigController extends Controller
{
    public function edit(Request $request): Response
    {
        $config = $request->user()->botConfig;

        return Inertia::render('bot/Setup', [
            'config' => $config ? [
                'smtp_host'         => $config->smtp_host,
                'smtp_port'         => $config->smtp_port,
                'smtp_username'     => $config->smtp_username,
                'smtp_security'     => $config->smtp_security,
                'smtp_from_name'    => $config->smtp_from_name,
                'smtp_from_address' => $config->smtp_from_address,
                'ai_provider'       => $config->ai_provider,
                'ai_api_url'        => $config->ai_api_url,
                'ai_model'          => $config->ai_model,
                'is_active'         => $config->is_active,
                // smtp_password and ai_api_key are never sent to frontend
            ] : null,
        ]);
    }

    public function update(Request $request): RedirectResponse
    {
        $data = $request->validate([
            'smtp_host'         => ['required', 'string', 'max:255'],
            'smtp_port'         => ['required', 'integer', 'min:1', 'max:65535'],
            'smtp_username'     => ['required', 'string', 'max:255'],
            'smtp_password'     => ['nullable', 'string', 'max:255'],
            'smtp_security'     => ['required', 'in:SSL,STARTTLS'],
            'smtp_from_name'    => ['required', 'string', 'max:255'],
            'smtp_from_address' => ['required', 'email', 'max:255'],
            'ai_provider'       => ['required', 'in:gemini,cloudflare'],
            'ai_api_url'        => ['nullable', 'string', 'max:512'],
            'ai_api_key'        => ['nullable', 'string', 'max:255'],
            'ai_model'          => ['nullable', 'string', 'max:128'],
            'is_active'         => ['boolean'],
        ]);

        $config = $request->user()->botConfig;

        if ($config) {
            if (empty($data['smtp_password'])) {
                unset($data['smtp_password']);
            }
            if (empty($data['ai_api_key'])) {
                unset($data['ai_api_key']);
            }
            $config->update($data);
        } else {
            $request->user()->botConfig()->create($data);
        }

        return back()->with('success', 'Konfiguration gespeichert.');
    }
}