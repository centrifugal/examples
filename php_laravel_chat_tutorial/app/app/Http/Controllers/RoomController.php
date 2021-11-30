<?php

namespace App\Http\Controllers;

use App\Models\Room;
use Exception;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\DB;
use Throwable;

class RoomController extends Controller
{
    public function index()
    {
        return view('rooms.index', [
            'rooms' => Room::with('users')->get()
        ]);
    }

    public function show(int $id)
    {
        return view('rooms.show', [
            'room' => Room::with(['users', 'messages'])->find($id)
        ]);
    }

    public function join(int $id)
    {
        $room = Room::find($id);
        $room->users()->attach(Auth::user()->id);

        return redirect()->route('rooms.show', $id);
    }

    public function store(Request $request)
    {
        $request->validate([
            'name' => ['required', 'string', 'max:32', 'unique:rooms'],
        ]);

        DB::beginTransaction();
        try {
            $room = Room::create(['name' => $request->get('name')]);
            $room->users()->attach(Auth::user()->id);
            DB::commit();
        } catch (Throwable $throwable) {
            DB::rollBack();

            throw new Exception('Error on create ');
        }

        return redirect('rooms');
    }
}
