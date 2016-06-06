# album
Server-side application for hosting home photo albums

# Motivation

Photos copied from camera's card to a home server become hard to come back to, because it's hard to organize them. Also, there's no way to rate, favorite and tag them without special software installed on the home computers accessing the shared library, not to mention phones and tablets.

I don't want to share my family photos with the "cloud" just because it's easier to do all those things. I'd like to be able to experience my private library in a way that's as easy and accessible as classic family albums, while adding some cool organizational features.

This is a side-project with me having not much time, so if you stumble upon it and want to help in any way possible, pull requests, issues, ideas welcome. Knowledge of Go or HTML, CSS, Javascript, and probably React (or Aurelia?), will be useful.

# Goals

- Starring (favoriting photos), rating, tagging
- Thumbnails for JPEGs and RAW files (CR2 in particular)
- Browsing all photos or only those matching certain criteria - be that search results, star, rating, tags
- Displaying photo metadata - shutter, aperture, iso, time taken, location (on a map perhaps?)
- Single `.dotfile` in photo directory root for thumbnails, metadata and index database
- Metadata stored as plain text or json files
- Make no other changes to photo directory structure
- Database used only for metadata index (for browsing performance), can be recreated
- Make use of filesystem change monitoring for realtime index updates
- Nice, web-based UI
- Provide realtime UI updates through websockets

# Plan (to do)
- ~~Photos folder crawler~~
- ~~Generate image file's SHA1 checksum~~
- ~~Generate thumbnails for JPEGs, skipping those already processed and unchanged~~
- Extract metadata from photos - date taken, GPS coordinates, basic photo parameters. If there is no EXIF data, photo file modification date is used as photo's date of origin
- Simple web UI for browsing photos (only JPEGs for now) - photos are organized by year/month/day by default
- Favorite/love a photo
- Use a key-value store for metadata index (boltdb?)
- Show favorited/loved photos
- Track files moved or detecting image file content change (for cases where e.g. the Windows photo viewer modifies the photo on rotation) - use hashes and metadata index; even though files might be moved or modified, their metadata will follow them
- Tag photo
- Show photos with given tag(s)
- Tag autocompletion
- Tag synonyms
- Thumbnails for RAW

# Nice to have
- Access full RAW files color resolution for preview, ability to adjust exposure
- GIF previews for movies (similar to what Google+ does) - will surely need an external library or tool for frame extraction
- Face tags
- Face recognition for autotagging (OpenCV includes algorithms for that)
- Rate photo?
- Show photos with at least/at most n stars?
- Contrast detection to find blurry or badly focused photos
- TBD
