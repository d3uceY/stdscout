port module Main exposing (main)

{-| StdScout — every executable on PATH, and the folders that keep them.

Two views over the same scan:

  - Applications: one row per executable, filterable, with the copy that
    actually runs marked and the shadowed duplicates labelled.
  - Locations: the PATH folders that hold a family of binaries, plus the ones
    that have gone missing.

Ports are the only way out: `request` asks main.js to call a Go binding,
`binariesFromJs` / `locationsFromJs` / `errorFromJs` bring the answers back.
Types and rendering live in Types and View.

-}

import Browser
import Types exposing (..)
import View



-- PORTS


port request : { action : String, target : String } -> Cmd msg


port binariesFromJs : (List Binary -> msg) -> Sub msg


port locationsFromJs : (List Location -> msg) -> Sub msg


port errorFromJs : (String -> msg) -> Sub msg



-- PROGRAM


init : () -> ( Model, Cmd Msg )
init _ =
    ( { tab = Applications
      , status = Loading
      , binaries = []
      , locations = []
      , query = ""
      , showSystem = False
      }
    , Cmd.batch
        [ request { action = "list", target = "" }
        , request { action = "locations", target = "" }
        ]
    )


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        ShowTab tab ->
            ( { model | tab = tab }, Cmd.none )

        ChangeQuery query ->
            ( { model | query = query }, Cmd.none )

        ToggleSystem showSystem ->
            ( { model | showSystem = showSystem }, Cmd.none )

        {- A folder is a way into the table, not a way out to Explorer. -}
        DrillInto path ->
            ( { model | tab = Applications, query = path }, Cmd.none )

        Reveal path ->
            ( model, request { action = "reveal", target = path } )

        GotBinaries binaries ->
            ( { model | binaries = binaries, status = Ready }, Cmd.none )

        GotLocations locations ->
            ( { model | locations = locations, status = Ready }, Cmd.none )

        GotError message ->
            ( { model | status = Failed message }, Cmd.none )


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.batch
        [ binariesFromJs GotBinaries
        , locationsFromJs GotLocations
        , errorFromJs GotError
        ]


main : Program () Model Msg
main =
    Browser.element
        { init = init
        , update = update
        , view = View.view
        , subscriptions = subscriptions
        }
