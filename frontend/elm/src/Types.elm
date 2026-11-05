module Types exposing
    ( Binary
    , Location
    , Model
    , Msg(..)
    , Status(..)
    , Tab(..)
    , countLabel
    , matchingBinaries
    , summary
    , visibleBinaries
    , visibleLocations
    )

{-| The records the ports carry, the model they land in, and the two filtered
views the table and grid render. No Html in here on purpose: the shape of the
data should be readable without a view in the way.
-}


type alias Binary =
    { name : String
    , path : String
    , logo : String
    , shadowed : Bool
    , system : Bool
    }


type alias Location =
    { path : String
    , label : String
    , logo : String
    , count : Int
    , system : Bool
    , dead : Bool
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
    , showSystem : Bool
    , version : String
    }


type Msg
    = ShowTab Tab
    | ChangeQuery String
    | ToggleSystem Bool
    | DrillInto String
    | Reveal String
    | GotBinaries (List Binary)
    | GotLocations (List Location)
    | GotError String



-- QUERIES


{-| What the search box matches, before the System filter is applied. The view
needs both numbers so it can tell "nothing matches" apart from "everything that
matches is a Windows binary you are hiding".
-}
matchingBinaries : Model -> List Binary
matchingBinaries model =
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


{-| What the table shows: Windows' own binaries are hidden until asked for,
then the search box narrows by name or path.
-}
visibleBinaries : Model -> List Binary
visibleBinaries model =
    List.filter (\binary -> model.showSystem || not binary.system) (matchingBinaries model)


visibleLocations : Model -> List Location
visibleLocations model =
    List.filter (\location -> model.showSystem || not location.system) model.locations


{-| Top-bar hint: how much of the scan is on screen, plus anything broken.
-}
summary : Model -> String
summary model =
    case model.tab of
        Locations ->
            let
                total =
                    String.fromInt (List.length model.locations)

                shown =
                    String.fromInt (List.length (visibleLocations model))

                dead =
                    List.length (List.filter .dead model.locations)
            in
            (if dead > 0 then
                String.fromInt dead ++ " missing · "

             else
                ""
            )
                ++ (if shown == total then
                        total ++ " folders"

                    else
                        shown ++ " of " ++ total ++ " folders"
                   )

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
                            List.length (visibleBinaries model)
                    in
                    if shown == total then
                        String.fromInt total ++ " on path"

                    else
                        String.fromInt shown ++ " of " ++ String.fromInt total


countLabel : Location -> String
countLabel location =
    if location.dead then
        "missing"

    else if location.count == 1 then
        "1 binary"

    else
        String.fromInt location.count ++ " binaries"
