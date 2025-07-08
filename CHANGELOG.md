# Changelog
Tracking from v1.3 and onwards if God Wills.


### v1.3:
- [x] Set `stream: true` for Mistral to allow live streaming of tokens.
- [x] Introduced live streaming in cli aswell (idk why im calling it livestreaming xd).
- [x] Introduced SSE in API `/ai/request` allow live streaming of tokens over HTTP.

### v1.4:
- [x] 80-120X FASTER! 
- [x] Introduced the Groq API (using model `meta-llama/llama-4-scout-17b-16e-instruct`). 
- [x] Introduced `ParseStreamedSSE` and commented out the older model for now, to allow live-streaming of tokens.
- [x] Changed the `chan string` to `<-chan AIResponse` in parser.go to allow more information to be processed (and also made the channel <- read-only).
- [x] 30 responses per minute now only, but better than the 60+ seconds that requests used to take.
- [x] **Reason I switched from Mistral:** Only allowed 1 request per 60-120 seconds, as I could only run 1 instance on my PC which is designed to handle one thread only.
