module View exposing (view)

{-| Everything the app draws. One function per region: the shell, the tabs, the
table, the folder grid.
-}

import Html exposing (Html, button, div, header, img, input, label, main_, nav, p, section, span, text)
import Html.Attributes as Attr
import Html.Events as Ev
import Types exposing (..)


view : Model -> Html Msg
view model =
    div [ Attr.class "shell" ]
        [ header [ Attr.class "topbar" ]
            [ div [ Attr.class "brand" ]
                [ brandSprite
                , span [ Attr.class "brand-name" ] [ text "StdScout" ]
                , release model.version
                ]
            , span [ Attr.class "topbar-meta" ] [ text (summary model) ]
            ]
        , nav [ Attr.class "tabs" ]
            [ div [ Attr.class "tabs-list" ]
                [ tabButton model.tab Applications "Applications"
                , tabButton model.tab Locations "Locations"
                ]
            , systemToggle model.showSystem
            ]
        , main_ [ Attr.class "content" ]
            (case model.tab of
                Applications ->
                    applicationsView model

                Locations ->
                    locationsView model
            )
        ]



-- CHROME


{-| An empty crop box. main.js mounts the sprite sheet player into it once, on
startup; the element is never rebuilt, so the player keeps its place across
re-renders.
-}
brandSprite : Html Msg
brandSprite =
    span [ Attr.class "brand-logo", Attr.attribute "aria-hidden" "true" ] []


{-| The release stamped into the binary at build time, in small print. Empty
in a browser preview, where nothing stamped it.
-}
release : String -> Html Msg
release version =
    if version == "" then
        text ""

    else
        span [ Attr.class "brand-version" ] [ text version ]


tabButton : Tab -> Tab -> String -> Html Msg
tabButton active tab labelText =
    button
        [ Attr.class
            ("tab"
                ++ (if active == tab then
                        " is-active"

                    else
                        ""
                   )
            )
        , Ev.onClick (ShowTab tab)
        ]
        [ text labelText ]


{-| Windows' own binaries are ~80% of a PATH, so they start hidden.
-}
systemToggle : Bool -> Html Msg
systemToggle showSystem =
    label
        [ Attr.class "toggle"
        , Attr.title "Include the executables Windows ships with"
        ]
        [ input
            [ Attr.type_ "checkbox"
            , Attr.checked showSystem
            , Ev.onCheck ToggleSystem
            ]
            []
        , span [ Attr.class "toggle-track" ] [ span [ Attr.class "toggle-thumb" ] [] ]
        , span [ Attr.class "toggle-text" ] [ text "System" ]
        ]



-- APPLICATIONS


applicationsView : Model -> List (Html Msg)
applicationsView model =
    [ input
        [ Attr.class "search"
        , Attr.type_ "search"
        , Attr.placeholder "Filter by name or path"
        , Attr.attribute "aria-label" "Filter applications by name or path"
        , Attr.autofocus True
        , Attr.value model.query
        , Ev.onInput ChangeQuery
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
                    visibleBinaries model
            in
            if List.isEmpty model.binaries then
                [ p [ Attr.class "empty" ] [ text "No executables on PATH." ] ]

            else if List.isEmpty visible then
                [ p [ Attr.class "empty" ] [ text (emptyReason model) ] ]

            else
                List.map row visible


{-| Say which empty it is: a filtered-out Windows folder is not the same as a
typo, and the fix is the toggle, not a new search.
-}
emptyReason : Model -> String
emptyReason model =
    let
        hidden =
            List.length (matchingBinaries model)
    in
    if hidden > 0 then
        String.fromInt hidden
            ++ (if hidden == 1 then
                    " Windows binary matches"

                else
                    " Windows binaries match"
               )
            ++ ". Turn on System to see them."

    else if String.trim model.query == "" then
        "Only Windows' own binaries are on PATH. Turn on System to see them."

    else
        "Nothing matches \"" ++ model.query ++ "\"."


row : Binary -> Html Msg
row binary =
    div [ Attr.class "row" ]
        [ span [ Attr.class "row-app" ]
            [ appIcon binary
            , span [ Attr.class "row-name", Attr.title binary.name ] [ text binary.name ]
            , shadowFlag binary
            ]
        , span [ Attr.class "row-path", Attr.title binary.path ] [ text binary.path ]
        , button
            [ Attr.class "row-action"
            , Attr.title ("Reveal " ++ binary.path)
            , Attr.attribute "aria-label" ("Reveal " ++ binary.path ++ " in its folder")
            , Ev.onClick (Reveal binary.path)
            ]
            [ img [ Attr.class "row-action-icon", Attr.src "/assets/folder.png", Attr.alt "" ] []
            ]
        ]


{-| The folder's brand sits over the file-type mark and a few pixels to its
left, so both marks stay readable: one says what the folder is, the other what
the file is.
-}
appIcon : Binary -> Html Msg
appIcon binary =
    span [ Attr.class "row-icon" ]
        [ brandMark "row-brand" binary.logo
        , img
            [ Attr.class "row-ext"
            , Attr.src ("/assets/ext-" ++ iconFor binary.name ++ ".png")
            , Attr.alt ""
            ]
            []
        ]


{-| The point of the list: typing this name runs an earlier copy, not this one.
-}
shadowFlag : Binary -> Html Msg
shadowFlag binary =
    if binary.shadowed then
        span
            [ Attr.class "row-flag"
            , Attr.title "An earlier folder on PATH has a copy, so typing this name runs that one"
            ]
            [ text "shadowed" ]

    else
        text ""


{-| One mark per executable type. Anything without art of its own falls back
to the generic application window.
-}
iconFor : String -> String
iconFor name =
    case String.toLower (Maybe.withDefault "" (List.head (List.reverse (String.split "." name)))) of
        "msc" ->
            "msc"

        "cpl" ->
            "cpl"

        "cmd" ->
            "cmd"

        "bat" ->
            "bat"

        "com" ->
            "com"

        "js" ->
            "js"

        "vbs" ->
            "vbs"

        "wsf" ->
            "wsf"

        _ ->
            "exe"



-- LOCATIONS


locationsView : Model -> List (Html Msg)
locationsView model =
    case model.status of
        Loading ->
            [ p [ Attr.class "empty" ] [ text "Scanning PATH…" ] ]

        Failed message ->
            [ p [ Attr.class "empty empty--error" ] [ text message ] ]

        Ready ->
            let
                visible =
                    visibleLocations model
            in
            if List.isEmpty visible then
                [ p [ Attr.class "empty" ]
                    [ text
                        (if List.isEmpty model.locations then
                            "No PATH folders."

                         else
                            "Only Windows' own folders are on PATH. Turn on System to see them."
                        )
                    ]
                ]

            else
                [ section [ Attr.class "grid" ] (List.map folder visible) ]


{-| A live folder drills into the table so you can see what is in it; a dead one
has nothing to show and says so instead of pretending to be clickable.
-}
folder : Location -> Html Msg
folder location =
    button
        (if location.dead then
            [ Attr.class "folder folder--dead"
            , Attr.title (location.path ++ " is on PATH but no longer exists")
            , Attr.disabled True
            ]

         else
            [ Attr.class "folder"
            , Attr.title (location.path ++ " — list its binaries")
            , Ev.onClick (DrillInto location.path)
            ]
        )
        [ span [ Attr.class "folder-art" ]
            [ img [ Attr.class "folder-base", Attr.src "/assets/folder.png", Attr.alt "" ] []
            , brandMark "folder-logo" location.logo
            ]
        , span [ Attr.class "folder-label" ] [ text location.label ]
        , span [ Attr.class "folder-count" ] [ text (countLabel location) ]
        ]


{-| The brand sticker sits on the folder's front panel. Empty logo = no sticker.
-}
brandMark : String -> String -> Html Msg
brandMark className slug =
    if slug == "" then
        text ""

    else
        img [ Attr.class className, Attr.src ("/logos/" ++ slug ++ ".svg"), Attr.alt "" ] []
