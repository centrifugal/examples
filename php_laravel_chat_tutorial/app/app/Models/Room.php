<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Room extends Model
{
    protected $fillable = [
        'name',
    ];

    public function users()
    {
        return $this->belongsToMany(User::class, 'users_rooms');
    }

    public function messages()
    {
        return $this->hasMany(Message::class);
    }
}
