port module Main exposing (main)

{-| StdScout — every executable on PATH, and the folders that keep them.

Two views over the same scan:

  - Applications: one row per executable, filterable.
  - Locations: the PATH folders that hold a family of binaries.

Ports are the only way out: `request` asks main.js to call a Go binding,
`binariesFromJs` / `locationsFromJs` / `errorFromJs` bring the answers back.

-}

import Browser
import Html exposing (Html, button, div, header, img, input, main_, nav, p, section, span, text)
import Html.Attributes as Attr
import Html.Events as Ev



-- PORTS


port request : { action : String, target : String } -> Cmd msg


port binariesFromJs : (List Binary -> msg) -> Sub msg


port locationsFromJs : (List Location -> msg) -> Sub msg


port errorFromJs : (String -> msg) -> Sub msg



-- TYPES


type alias Binary =
    { name : String
    , path : String
    }


type alias Location =
    { path : String
    , label : String
    , logo : String
    , count : Int
    }


type Tab
    = Applications
    | Locations


type Status
    = Loading
    | Ready
    | Failed String


type alias Model =
    { tab : Tab
    , status : Status
    , binaries : List Binary
    , locations : List Location
    , query : String
    }


type Msg
    = ShowTab Tab
    | Reveal String
    | QueryChanged String
    | GotBinaries (List Binary)
    | GotLocations (List Location)
    | GotError String



-- INIT / UPDATE


init : () -> ( Model, Cmd Msg )
init _ =
    ( { tab = Applications, status = Loading, binaries = [], locations = [], query = "" }
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

        Reveal path ->
            ( model, request { action = "reveal", target = path } )

        QueryChanged query ->
            ( { model | query = query }, Cmd.none )

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



-- VIEW


view : Model -> Html Msg
view model =
    div [ Attr.class "shell" ]
        [ header [ Attr.class "topbar" ]
            [ div [ Attr.class "brand" ]
                [ img [ Attr.class "brand-logo", Attr.src "/logo.png", Attr.alt "" ] []
                , span [ Attr.class "brand-name" ] [ text "StdScout" ]
                ]
            , span [ Attr.class "topbar-meta" ] [ text (summary model) ]
            ]
        , nav [ Attr.class "tabs" ]
            [ tabButton model.tab Applications "Applications"
            , tabButton model.tab Locations "Locations"
            ]
        , main_ [ Attr.class "content" ]
            (case model.tab of
                Applications ->
                    applicationsView model

                Locations ->
                    locationsView model
            )
        ]


tabButton : Tab -> Tab -> String -> Html Msg
tabButton active tab label =
    button
        [ Attr.class ("tab" ++ (if active == tab then " is-active" else ""))
        , Ev.onClick (ShowTab tab)
        ]
        [ text label ]


applicationsView : Model -> List (Html Msg)
applicationsView model =
    [ input
        [ Attr.class "search"
        , Attr.type_ "search"
        , Attr.placeholder "Filter by name or path"
        , Attr.attribute "aria-label" "Filter applications by name or path"
        , Attr.autofocus True
        , Attr.value model.query
        , Ev.onInput QueryChanged
        ]
        []
    , section [ Attr.class "list" ]
        [ div [ Attr.class "list-head" ]
            [ span [] [ text "Application" ]
            , span [] [ text "Path" ]
            , span [] []
            ]
        , div [ Attr.class "list-body" ] (rows model)
        ]
    ]


rows : Model -> List (Html Msg)
rows model =
    case model.status of
        Loading ->
            [ p [ Attr.class "empty" ] [ text "Scanning PATH…" ] ]

        Failed message ->
            [ p [ Attr.class "empty empty--error" ] [ text message ] ]

        Ready ->
            let
                visible =
                    filtered model
            in
            if List.isEmpty model.binaries then
                [ p [ Attr.class "empty" ] [ text "No executables on PATH." ] ]

            else if List.isEmpty visible then
                [ p [ Attr.class "empty" ] [ text ("Nothing matches \"" ++ model.query ++ "\".") ] ]

            else
                List.map row visible


row : Binary -> Html Msg
row binary =
    div [ Attr.class "row" ]
        [ span [ Attr.class "row-app" ]
            [ img
                [ Attr.class "row-icon"
                , Attr.src ("/assets/ext-" ++ extClass binary.name ++ ".png")
                , Attr.alt ""
                ]
                []
            , span [ Attr.class "row-name", Attr.title binary.name ] [ text binary.name ]
            ]
        , span [ Attr.class "row-path", Attr.title binary.path ] [ text binary.path ]
        , button
            [ Attr.class "row-action"
            , Attr.title ("Reveal " ++ binary.path)
            , Ev.onClick (Reveal binary.path)
            ]
            [ text "Open folder" ]
        ]


locationsView : Model -> List (Html Msg)
locationsView model =
    case model.status of
        Loading ->
            [ p [ Attr.class "empty" ] [ text "Scanning PATH…" ] ]

        Failed message ->
            [ p [ Attr.class "empty empty--error" ] [ text message ] ]

        Ready ->
            if List.isEmpty model.locations then
                [ p [ Attr.class "empty" ] [ text "No PATH folders." ] ]

            else
                [ section [ Attr.class "grid" ] (List.map folder model.locations) ]


folder : Location -> Html Msg
folder location =
    button
        [ Attr.class "folder"
        , Attr.title location.path
        , Ev.onClick (Reveal location.path)
        ]
        [ span [ Attr.class "folder-art" ]
            [ img [ Attr.class "folder-base", Attr.src "/assets/folder.png", Attr.alt "" ] []
            , brandMark location.logo
            ]
        , span [ Attr.class "folder-label" ] [ text location.label ]
        , span [ Attr.class "folder-count" ] [ text (String.fromInt location.count ++ " binaries") ]
        ]


{-| The brand sticker sits on the folder's front panel. Empty logo = no sticker.
-}
brandMark : String -> Html Msg
brandMark slug =
    if slug == "" then
        text ""

    else
        img [ Attr.class "folder-logo", Attr.src ("/logos/" ++ slug ++ ".svg"), Attr.alt "" ] []


{-| Rows group extensions into three marks: real executables, scripts, and the
Windows management snap-ins (.msc / .cpl).
-}
extClass : String -> String
extClass name =
    case String.toLower (Maybe.withDefault "" (List.head (List.reverse (String.split "." name)))) of
        "msc" ->
            "system"

        "cpl" ->
            "system"

        "cmd" ->
            "script"

        "bat" ->
            "script"

        "com" ->
            "script"

        "js" ->
            "script"

        "vbs" ->
            "script"

        "wsf" ->
            "script"

        _ ->
            "exe"


filtered : Model -> List Binary
filtered model =
    let
        needle =
            String.toLower (String.trim model.query)
    in
    if needle == "" then
        model.binaries

    else
        List.filter
            (\binary -> String.contains needle (String.toLower (binary.name ++ " " ++ binary.path)))
            model.binaries


summary : Model -> String
summary model =
    case model.tab of
        Locations ->
            String.fromInt (List.length model.locations) ++ " folders"

        Applications ->
            case model.status of
                Loading ->
                    "scanning…"

                Failed _ ->
                    "unavailable"

                Ready ->
                    let
                        total =
                            List.length model.binaries

                        shown =
                            List.length (filtered model)
                    in
                    if shown == total then
                        String.fromInt total ++ " on path"

                    else
                        String.fromInt shown ++ " of " ++ String.fromInt total



-- PROGRAM


main : Program () Model Msg
main =
    Browser.element
        { init = init
        , update = update
        , view = view
        , subscriptions = subscriptions
        }
