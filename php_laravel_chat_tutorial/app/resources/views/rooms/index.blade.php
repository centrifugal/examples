<x-app-layout>
    <x-slot name="header">
        <h2 class="font-semibold text-xl text-gray-800 leading-tight">
            {{ __('Dashboard') }}
        </h2>
    </x-slot>

    <div class="py-12">
        <div class="max-w-7xl mx-auto sm:px-6 lg:px-8">
            <form class="my-5" method="post" action="{{ route('rooms.store') }}">
                @csrf
                <div>
                    <x-input class="block mt-1 w-full" type="text" name="name" required autofocus />
                </div>
                <div class="flex items-center justify-end mt-4">
                    <x-button>
                        {{ __('Add room') }}
                    </x-button>
                </div>
            </form>

            <div class="bg-white overflow-hidden shadow-sm sm:rounded-lg">
                <div class="p-6 bg-white border-b border-gray-200">
                    Rooms:
                </div>

                @foreach($rooms as $room)
                    <div class="my-2 ml-5">
                        {{ $room->name }}
                        @if ($room->users->where('id', '=', Auth::user()->id)->first())
                            <a href="{{ route('rooms.show', $room->id) }}" class="inline-block px-4 py-2 bg-green-500 rounded-md text-xs text-white uppercase hover:bg-green-300">
                                View
                            </a>
                        @else
                            <form class="inline-block px-4 py-2 bg-blue-700 rounded-md text-xs text-white hover:bg-blue-500" method="post" action="{{ route('rooms.join', $room->id) }}">
                                @csrf
                                <button type="submit">JOIN</button>
                            </form>
                        @endif
                    </div>
                @endforeach
            </div>
        </div>
    </div>
</x-app-layout>
