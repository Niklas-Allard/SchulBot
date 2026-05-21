<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('bot_configs', function (Blueprint $table) {
            $table->id();
            $table->foreignId('user_id')->constrained()->cascadeOnDelete();

            // IMAP
            $table->string('imap_host');
            $table->unsignedSmallInteger('imap_port')->default(993);
            $table->string('imap_username');
            $table->text('imap_password');
            $table->string('imap_security', 16)->default('SSL');
            $table->string('imap_mailbox')->default('INBOX');

            // SMTP
            $table->string('smtp_host');
            $table->unsignedSmallInteger('smtp_port')->default(465);
            $table->string('smtp_username');
            $table->text('smtp_password');
            $table->string('smtp_security', 16)->default('SSL');
            $table->string('smtp_from_name');
            $table->string('smtp_from_address');

            // AI
            $table->string('ai_provider', 32)->default('gemini');
            $table->string('ai_api_url')->nullable();
            $table->text('ai_api_key')->nullable();
            $table->string('ai_model')->nullable();

            $table->boolean('is_active')->default(true);
            $table->timestamps();
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('bot_configs');
    }
};
