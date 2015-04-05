# album
Server-side application for hosting home photo albums

# Motivation

Photos copied from camera's card to a home server become hard to come back to, because it's hard to organize them, there's no way to rate, star and tag them without special software installed on the home computers accessing them, not to mention phones and tablets.

I don't want to share my family photos with the "cloud" just because it's easier to do all those things. I'd like to be able to experience them in a way that's as close to browsing the classic family albums as possible, while getting access to some great possibilities technology gives us.

This is a side-project with me having not much time, so if you stumble upon that project and want to help in any way possible, that would be great. Pull requests, issues, ideas welcome. Knowledge of Go and html, css, javascript, react(?) will be useful.

# Goals

- Starring (favoriting photos), rating, tagging
- Thumbnails for JPEGs and RAW files (CR2 in particular)
- Browsing all photos or only those matching certain criteria - be that search results, star, rating, tags
- Displaying photo metadata - shutter, aperture, iso, time taken, location (on a map perhaps?)
- Single `.dotfile` in photo directory root for thumbnails, metadata and index database
- Metadata stored as plain text or json files
- Making no other changes to photo directory structure
- Database used only for metadata index (for browsing performance), can be recreated
- Nice, web-based UI
- Make use of filesystem change monitoring for realtime index updates and providing instant UI updates through websockets

# Plan (to do)

- Photos folder crawling
- Generate image file's SHA1 checksum for tracking files moved or detecting image file content change (for cases where e.g. the Windows photo viewer modifies the photo on rotation)
- Generating thumbnails for JPEGs, skipping those already processed and unchanged
- Simple web UI for browsing photos (only JPEGs for now)
- Star photo
- Use a key-value store for metadata index
- Show starred photos
- Tag photo
- Show photos with given tag(s)
- Star photo
- Show photos with at least n stars
- Thumbnails for CR2
- TBD
