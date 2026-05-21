<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Support\Facades\Crypt;

class BotConfig extends Model
{
    protected $fillable = [
        'user_id',
        'smtp_host', 'smtp_port', 'smtp_username', 'smtp_password', 'smtp_security', 'smtp_from_name', 'smtp_from_address',
        'ai_provider', 'ai_api_url', 'ai_api_key', 'ai_model',
        'is_active',
    ];

    protected $hidden = ['smtp_password', 'ai_api_key'];

    protected function casts(): array
    {
        return [
            'smtp_port' => 'integer',
            'is_active' => 'boolean',
        ];
    }

    public function user(): BelongsTo
    {
        return $this->belongsTo(User::class);
    }

    public function setSmtpPasswordAttribute(string $value): void
    {
        $this->attributes['smtp_password'] = Crypt::encryptString($value);
    }

    public function setAiApiKeyAttribute(?string $value): void
    {
        $this->attributes['ai_api_key'] = $value ? Crypt::encryptString($value) : null;
    }

    public function getSmtpPasswordDecrypted(): string
    {
        return Crypt::decryptString($this->attributes['smtp_password']);
    }

    public function getAiApiKeyDecrypted(): ?string
    {
        return $this->attributes['ai_api_key']
            ? Crypt::decryptString($this->attributes['ai_api_key'])
            : null;
    }
}