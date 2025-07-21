# CHANGELOG

# unreleased
- Refactorings to add a unique number to the name and to correct the id's. Plus, a rename of the function to HugePayloadHandler to support more functionality in the future. Plus a new CHANGELOG and a modified readme
- Added plugin architecture for payload handlers (PayloadPlugin interface)
- Added /stream_payload endpoint for streaming large JSON arrays
- Refactored StreamingPayloadHandler into its own file
- Improved documentation in README.md and handler files
- Added tests for streaming payload handler
- initial commit
