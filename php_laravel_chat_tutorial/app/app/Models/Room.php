<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\HasMany;

class Room extends Model
{
    protected $fillable = [
        'name',
    ];

    /*public function userRooms(): HasMany
    {
        return $this->hasMany(UserRoom::class, 'room_id');
    }*/

    public function users()
    {
        return $this->belongsToMany(User::class, 'users_rooms');
    }
}
