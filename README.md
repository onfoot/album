# album
Server-side application for hosting home photo albums

# Motivation

Photos copied from camera's card to a home server become hard to come back to, because it's hard to organize them. Also, there's no way to rate, favorite and tag them without special software installed on the home computers accessing them a shared library, not to mention phones and tablets.

I don't want to share my family photos with the "cloud" just because it's easier to do all those things. I'd like to be able to experience my private library in a way that's as easy and accessible as classic family albums, while adding some cool organizational features.

This is a side-project with me having not much time, so if you stumble upon it and want to help in any way possible, pull requests, issues, ideas welcome. Knowledge of Go or html, css, javascript, and probably react, will be useful.

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

- Photos folder crawler
- Generate image file's SHA1 checksum for tracking files moved or detecting image file content change (for cases where e.g. the Windows photo viewer modifies the photo on rotation)
- Generate thumbnails for JPEGs, skipping those already processed and unchanged
- Simple web UI for browsing photos (only JPEGs for now)
- Favorite photo
- Use a key-value store for metadata index
- Show favorite photos
- Tag photo
- Show photos with given tag(s)
- Rate photo
- Show photos with at least n stars
- Thumbnails for CR2
- TBD
