port module Main exposing (main)

{-| stdscout UI.

Elm owns the DOM. Everything outside it is reached through ports:

  - `command` (out) asks main.js to call a Go binding.
  - `animate` (out) asks main.js to run a GSAP tween.
  - `stateFromJs` (in) receives the snapshot Go returned.
  - `logFromJs` (in) receives activity from Go bindings and Go events.

Elm never imports a Wails module, and main.js never builds DOM. That split is
the whole point of this skeleton.

-}

import Browser
import Html exposing (Html, button, div, footer, h1, header, main_, p, section, span, text)
import Html.Attributes as Attr
import Html.Events as Ev



-- PORTS


port command : { action : String, amount : Int } -> Cmd msg


port animate : String -> Cmd msg


port stateFromJs : (CounterState -> msg) -> Sub msg


port logFromJs : (LogEntry -> msg) -> Sub msg


{-| Shape of the snapshot returned by CounterService. Must stay a record of
primitives so it can cross the port boundary.
-}
type alias CounterState =
    { count : Int
    , lastAction : String
    , updatedAt : String
    }


type alias LogEntry =
    { at : String
    , level : String
    , message : String
    }


type alias Flags =
    { startedAt : String }



-- MODEL


type alias Model =
    { count : Int
    , lastAction : String
    , updatedAt : String
    , log : List LogEntry
    , pending : Bool
    }


init : Flags -> ( Model, Cmd Msg )
init flags =
    ( { count = 0
      , lastAction = "boot"
      , updatedAt = flags.startedAt
      , log = []
      , pending = True
      }
    , command { action = "sync", amount = 0 }
    )



-- UPDATE


type Msg
    = Requested String Int
    | GotState CounterState
    | GotLog LogEntry


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        Requested action amount ->
            ( { model | pending = True }
            , command { action = action, amount = amount }
            )

        GotState state ->
            ( { model
                | count = state.count
                , lastAction = state.lastAction
                , updatedAt = state.updatedAt
                , pending = False
              }
            , animate "count"
            )

        GotLog entry ->
            ( { model | log = List.take 6 (entry :: model.log) }
            , animate "log"
            )


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.batch
        [ stateFromJs GotState
        , logFromJs GotLog
        ]



-- VIEW


view : Model -> Html Msg
view model =
    div [ Attr.class "app" ]
        [ header [ Attr.class "topbar" ]
            [ div [ Attr.class "brand" ]
                [ span [ Attr.class "brand-mark" ] [ text "◎" ]
                , h1 [ Attr.class "brand-name" ] [ text "stdscout" ]
                , span [ Attr.class "brand-tag" ] [ text "elm · ports · go" ]
                ]
            , div [ Attr.class ("status" ++ busyClass model.pending) ]
                [ span [ Attr.class "status-dot" ] []
                , span [] [ text (statusText model.pending) ]
                ]
            ]
        , main_ [ Attr.class "stage" ]
            [ section [ Attr.class "panel panel--counter" ]
                [ p [ Attr.class "label" ] [ text "counter · owned by go" ]
                , span [ Attr.class "value", Attr.id "count" ] [ text (String.fromInt model.count) ]
                , div [ Attr.class "row" ]
                    [ button [ Attr.class "btn", Ev.onClick (Requested "decrement" 1) ] [ text "−1" ]
                    , button [ Attr.class "btn btn--primary", Ev.onClick (Requested "increment" 1) ] [ text "+1" ]
                    , button [ Attr.class "btn", Ev.onClick (Requested "increment" 10) ] [ text "+10" ]
                    , button [ Attr.class "btn btn--ghost", Ev.onClick (Requested "reset" 0) ] [ text "reset" ]
                    ]
                , p [ Attr.class "meta" ]
                    [ text ("last " ++ model.lastAction ++ " · " ++ model.updatedAt) ]
                ]
            , section [ Attr.class "panel panel--log", Attr.id "log" ]
                [ p [ Attr.class "label" ] [ text "event stream" ]
                , div [ Attr.class "log-list" ]
                    (if List.isEmpty model.log then
                        [ p [ Attr.class "log-empty" ] [ text "waiting for go…" ] ]

                     else
                        List.map logRow model.log
                    )
                ]
            ]
        , footer [ Attr.class "footer" ]
            [ span [] [ text "elm" ]
            , span [ Attr.class "arrow" ] [ text "→" ]
            , span [] [ text "ports" ]
            , span [ Attr.class "arrow" ] [ text "→" ]
            , span [] [ text "js" ]
            , span [ Attr.class "arrow" ] [ text "→" ]
            , span [] [ text "bindings" ]
            , span [ Attr.class "arrow" ] [ text "→" ]
            , span [ Attr.class "footer-go" ] [ text "go" ]
            ]
        ]


logRow : LogEntry -> Html Msg
logRow entry =
    div [ Attr.class ("log-row log-row--" ++ entry.level) ]
        [ span [ Attr.class "log-time" ] [ text entry.at ]
        , span [ Attr.class "log-msg" ] [ text entry.message ]
        ]


busyClass : Bool -> String
busyClass pending =
    if pending then
        " is-busy"

    else
        ""


statusText : Bool -> String
statusText pending =
    if pending then
        "waiting on go"

    else
        "linked"



-- PROGRAM


main : Program Flags Model Msg
main =
    Browser.element
        { init = init
        , update = update
        , view = view
        , subscriptions = subscriptions
        }
