<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Message extends Model
{
    protected $fillable = [
        'message',
        'room_id',
        'sender_id',
    ];

    public function room()
    {
        return $this->hasOne(Room::class);
    }

    public function user()
    {
        return $this->belongsTo(User::class, 'sender_id');
    }
}
