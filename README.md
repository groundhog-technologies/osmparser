# OSM data parser

Converts OSM PBF file to Element.


## GeoJSON

All data is given as Feature Collection.


### OSM element to GeoJSON transform rule

- `Node`:
    - `GeometryPoint`

- `Way`:
    - If `area=no`:
        - `GeometryLineString`

    - If `area` is yes:
        - If `highway=*` or `barrier=*`:
            - `GeometryLineString`
        - Else:
            - `GeometryPolygon`

- `Relation`:

    - If `type=multipolygon`:
        - GeometryMultipolygon

    - Else `type=*` or no type:
        - `GeometryCollection`

    - Note:
        - Remove recursive relationmember.
