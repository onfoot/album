# album
Server-side application for hosting home photo albums

# Motivation

Photos copied from camera's card to a home server become hard to come back to, because it's hard to organize them. Also, there's no way to rate, favorite and tag them without special software installed on the home computers accessing the shared library, not to mention phones and tablets.

I don't want to share my family photos with the "cloud" just because it's easier to do all those things. I'd like to be able to experience my private library in a way that's as easy and accessible as classic family albums, while adding some cool organizational features.

This is a side-project with me having not much time, so if you stumble upon it and want to help in any way possible, pull requests, issues, ideas welcome. Knowledge of Go or HTML, CSS, Elm or React (that's still undetermined - anything single-page-web-"app"-capable is fine), will be useful.

# Goals

- Favoriting photos, rating, tagging
- Thumbnails for JPEGs and RAW files
- Browsing all photos or only those matching certain criteria - be that search results, favorite, rating, tags
- Displaying photo metadata - shutter, aperture, iso, time taken, location (on a map perhaps? one not requiring an per-app-API key would be nice)
- Single `.dotfile` in photo directory root for thumbnails, metadata and index database
- Metadata stored as plain text or json files
- Make no other changes to photo directory structure
- Database used only for metadata index (for browsing performance), can be recreated
- Make use of filesystem change monitoring for realtime index updates
- Nice, web-based UI
- Provide realtime UI updates through websockets

# Plan (TODOs and ideas)

- ~~Photos folder crawler~~
- ~~Generate image file's SHA1 checksum~~
- ~~Generate thumbnails for JPEGs, skipping those already processed and unchanged~~
- Extract metadata from photos - date taken, GPS coordinates, basic photo parameters. If there is no EXIF data, photo file modification date is used as date of origin
- Simple web UI for browsing photos (only JPEGs for now, although it could generate JPEGs for RAWs as well, given a good external tool to convert those) - photos are organized by year/month/day, by default
- Watch directory for additions and changes, and update the index - lossy for now, i.e. if photo is moved, its additional metadata is lost. As there's no additional metadata for now, we're fine until we…
- Favorite a photo (by attaching a plain text or json file in metadata directory)
- Use a discardable, though permanent, key-value store for metadata index (boltdb?)
- Show favorited photos
- Track files moved - if there's a hash that doesn't have a file that points to it any longer, find out what happened to it - actually, if we index a new file with an apparently existing hash, we found the original. Also, original file's hash is metadata ID
- Detect image file content change (for cases where e.g. a photo viewer modifies the photo after performing edits, like rotation, even the lossless rotation that JPEG allows) - if the same filename has a different hash, figure out the change
- Tag photo
- Show photos with given tag(s)
- Tag autocompletion
- Tag synonyms
- Thumbnails for RAW

# Nice to have
- Detect bitrot? Even though the modification dates are not updated, checksums are different
- Pick a sensible `today` time slot to group together series of photos snapped at a late-night party for example
- Access full RAW files color resolution for preview, ability to adjust exposure
- low-fps mp4 animated previews for videos (similar to what Google+ or YouTube does) - will surely need an external library or tool for frame extraction - avconv/ffmpeg could generate 10-frame-per-video movies with no sound that would be automatically played by the web browser in loop
- Face tags
- Face recognition for autotagging (OpenCV includes algorithms for that)
- Rate photo?
- Show photos with at least/at most n stars?
- Contrast detection to find blurry or badly focused photos
- TBD
