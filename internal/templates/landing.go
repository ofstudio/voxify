package templates

import (
	"html/template"

	"github.com/ofstudio/voxify/internal/domain"
)

// LandingData is the input context for LandingTemplate
type LandingData struct {
	Feed     domain.FeedInfo
	Episodes []*domain.Episode
}

// LandingTemplate - template for the podcast landing page.
var LandingTemplate *template.Template

// landingTemplateHTML - HTML template source for the landing page.
const landingTemplateHTML =
// language=GoTemplate
`<!doctype html>
<html lang="ru">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
	<meta name="keywords" content="{{ .Feed.Keywords }}">
    <title>{{ .Feed.Title }}</title>
    <meta name="color-scheme" content="light dark">
	{{if .Episodes }}
    {{ if .Feed.RSSLink }}<link rel="alternate" type="application/rss+xml" title="Podcast RSS" href="{{ .Feed.RSSLink }}">{{ end }}
	{{end}}
    <style>
        :root {
            --bg: #f6f7fb;
            --elev-1: #ffffff;
            --text: #0b0c0f;
            --muted: #5f6472;
            --link: #0b68ff;
            --chip-bg: #eef1f8;
            --border-soft: rgba(12, 14, 22, 0.08);
            --shadow-1: 0 1px 2px rgba(0, 0, 0, .06), 0 8px 24px rgba(0, 0, 0, .08);
            --shadow-2: 0 2px 6px rgba(0, 0, 0, .08), 0 16px 36px rgba(0, 0, 0, .12);
            --btn-bg: linear-gradient(180deg, #ff9f0a, #ff7a00);
            --btn-text: #fff;
            --rss: #ff7a00;
            --ok: #12b886;
            --maxw: 1040px;
            --radius: 18px;
        }

        @media (prefers-color-scheme: dark) {
            :root {
                --bg: #0c0e16;
                --elev-1: #141824;
                --text: #f1f3f5;
                --muted: #a4a9b6;
                --link: #7ab5ff;
                --chip-bg: #1b2030;
                --border-soft: rgba(255, 255, 255, 0.06);
                --shadow-1: 0 1px 2px rgba(0, 0, 0, .6), 0 8px 24px rgba(0, 0, 0, .45);
                --shadow-2: 0 2px 8px rgba(0, 0, 0, .55), 0 22px 48px rgba(0, 0, 0, .6);
                --btn-bg: linear-gradient(180deg, #ffb340, #ff7a00);
                --btn-text: #121212;
            }
        }

        html, body {
            height: 100%;
        }

        body {
            margin: 0;
            font: 16px/1.5 -apple-system, BlinkMacSystemFont, Segoe UI, Roboto, Helvetica, Arial, "Apple Color Emoji", "Segoe UI Emoji";
            background: var(--bg);
            color: var(--text);
            -webkit-font-smoothing: antialiased;
            -moz-osx-font-smoothing: grayscale;
        }

        .wrap {
            max-width: var(--maxw);
            margin: 0 auto;
            padding: 24px;
        }

        /* Header card */
        .hero {
            background: var(--elev-1);
            border-radius: var(--radius);
            box-shadow: var(--shadow-2);
            padding: clamp(16px, 2.5vw, 28px);
            display: grid;
            grid-template-columns: 180px 1fr;
            gap: clamp(16px, 2.2vw, 28px);
            align-items: center;
        }

        .cover {
            width: 100%;
            aspect-ratio: 1 / 1;
            border-radius: 16px;
            overflow: hidden;
            box-shadow: var(--shadow-1);
            background: #ddd;
        }

        .cover img {
            width: 100%;
            height: 100%;
            object-fit: cover;
            display: block;
        }

        .title {
            margin: 0 0 6px 0;
            font-size: clamp(22px, 3.2vw, 36px);
            font-weight: 700;
            letter-spacing: -0.02em;
        }

        .meta {
            color: var(--muted);
            font-size: 14px;
            display: flex;
            gap: 10px;
            flex-wrap: wrap;
        }

        .meta a {
            color: var(--link);
            text-decoration: none;
        }

        .meta a:hover {
            text-decoration: underline;
        }

        .desc {
            margin: 14px 0 16px;
            color: var(--text);
            opacity: .92;
        }

        .chips {
            display: flex;
            flex-wrap: wrap;
            gap: 8px;
            margin: 8px 0 18px;
        }

        .chip {
            background: var(--chip-bg);
            color: var(--muted);
            padding: 6px 10px;
            border-radius: 999px;
            font-size: 12px;
            border: 1px solid var(--border-soft);
        }

        .actions {
            display: flex;
            gap: 12px;
            flex-wrap: wrap;
        }

        .btn {
            display: inline-flex;
            align-items: center;
            gap: 10px;
            padding: 10px 14px;
            border-radius: 12px;
            text-decoration: none;
            color: var(--btn-text);
            background: var(--btn-bg);
            box-shadow: 0 6px 18px rgba(255, 122, 0, .35), inset 0 1px rgba(255, 255, 255, .25);
            font-weight: 600;
            transition: transform .08s ease, filter .2s ease;
        }

        .btn:active {
            transform: translateY(1px);
        }

        .btn .rss {
            width: 16px;
            height: 16px;
        }

        /* Episodes */
        .episodes {
            margin: 26px 0 14px;
            display: grid;
            grid-template-columns: 1fr;
            gap: 16px;
        }

        h2.no-content {
            color: var(--muted);
            font-size: 24px;
            font-weight: 600;
            text-align: center;
            padding: 32px 0 0;
        }

        p.no-content {
            color: var(--muted);
            font-size: 16px;
            text-align: center;
            margin: 0 0 64px;
        }

        @media (min-width: 760px) {
            .episodes {
                grid-template-columns: 1fr 1fr;
            }
        }

        .card {
            background: var(--elev-1);
            border-radius: 16px;
            box-shadow: var(--shadow-1);
            padding: 12px;
            display: grid;
            grid-template-columns: 110px 1fr;
            gap: 14px;
            text-decoration: none;
            color: inherit;
            transition: transform .06s ease, box-shadow .2s ease;
            border: 1px solid var(--border-soft);
        }

        .card:hover {
            box-shadow: var(--shadow-2);
        }

        .thumb {
            width: 100%;
            aspect-ratio: 1/1;
            border-radius: 12px;
            overflow: hidden;
            background: #ccc;
        }

        .thumb img {
            width: 100%;
            height: 100%;
            object-fit: cover;
            display: block;
        }

        .card h3 {
            margin: 2px 0 6px;
            font-size: 16px;
            line-height: 1.35;
        }

        .card p {
            margin: 0 0 8px;
            color: var(--muted);
            font-size: 14px;
        }

        .card .row {
            display: flex;
            gap: 10px;
            color: var(--muted);
            font-size: 12px;
        }

        /* Footer */
        footer {
            margin: 28px 0 40px;
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 12px;
            flex-wrap: wrap;
        }

        .brand {
            display: inline-flex;
            align-items: center;
            gap: 10px;
            color: var(--muted);
            text-decoration: none;
        }

        .brand svg {
            width: 28px;
            height: 28px;
            display: block;
        }

        /* Layout tweaks */
        @media (max-width: 680px) {
            .hero {
                grid-template-columns: 1fr;
                padding: 0; /* full-bleed cover */
                gap: 0; /* remove inner gap so the cover touches edges */
            }

            .hero > .cover {
                border-radius: var(--radius) var(--radius) 0 0;
            }

            .hero > :not(.cover) {
                padding: clamp(16px, 2.5vw, 28px);
            }

            .actions .btn {
                width: 100%;
                justify-content: center;
            }
        }

        /* Desktop cover size bump */
        @media (min-width: 1024px) {
            .hero {
                grid-template-columns: 260px 1fr;
            }
        }
    </style>
</head>
<body>
<div class="wrap">
    <section class="hero" aria-labelledby="podcast-title">
        <div class="cover" aria-hidden="true">
            {{if .Feed.ImageUrl }}
                <img src="{{ .Feed.ImageUrl }}" alt="Artwork" loading="eager">
            {{ end }}
        </div>
        <div>
            <h1 id="podcast-title" class="title">{{ .Feed.Title }}</h1>
            <div class="meta">
                {{ if .Feed.Author }}<span>By <strong>{{ .Feed.Author }}</strong></span>{{ end }}
                {{ if .Feed.WebsiteLink }}
                    <span>·</span><a href="{{ .Feed.WebsiteLink }}" target="_blank" rel="noopener">Website</a>{{ end }}
            </div>
            <p class="desc">{{ .Feed.Description }}</p>
            {{ $categories := categoriesList .Feed.Categories }}
            {{ if $categories }}
                <div class="chips" aria-label="categories">
                    {{ range $categories }}<span class="chip">{{ . }}</span>{{ end }}
                </div>
            {{ end }}
            <div class="actions">
				{{if .Episodes }}
                {{ if .Feed.RSSLink }}
                    <a class="btn" href="{{ .Feed.RSSLink }}">
                        <svg class="rss" viewBox="0 0 24 24" aria-hidden="true">
                            <path d="M3.36247 17.2751C4.25425 17.2751 5.10951 17.6293 5.74009 18.2599C6.37068 18.8905 6.72494 19.7457 6.72494 20.6375C6.72494 22.4576 5.21337 24 3.36247 24C1.54242 24 0 22.4576 0 20.6375C0 19.7457 0.354259 18.8905 0.984844 18.2599C1.61543 17.6293 2.47069 17.2751 3.36247 17.2751ZM0 0C6.3652 0 12.4697 2.52856 16.9706 7.02944C21.4714 11.5303 24 17.6348 24 24H19.635C19.635 18.7925 17.5663 13.7983 13.884 10.116C10.2017 6.43372 5.20752 4.36504 0 4.36504V0ZM0 8.73008C4.04984 8.73008 7.9338 10.3389 10.7975 13.2025C13.6611 16.0662 15.2699 19.9502 15.2699 24H10.9049C10.9049 21.1078 9.75598 18.3341 7.71092 16.2891C5.66585 14.244 2.89216 13.0951 0 13.0951V8.73008Z"
                                  fill="currentColor"/>
                        </svg>
                        RSS feed
                    </a>
                {{ end }}
				{{ end }}
            </div>
        </div>
    </section>

    <section aria-labelledby="episodes-title" style="margin-top: 22px;">
        {{ if eq (len .Episodes) 0 }}
            <h2 class="no-content">No episodes yet</h2>
            <p class="no-content">Add some videos using Voxify telegram bot.</p>
        {{ else }}
            <h2 id="episodes-title" style="position:absolute; left:-9999px;">Latest episodes</h2>
            <div class="episodes">
                {{ range $i, $e := .Episodes }}
                    {{ if lt $i 10 }}
                        <a class="card" href="{{ episodeURL $e }}" target="_blank" rel="noopener">
                            <div class="thumb">{{if $e.ThumbnailFile}}<img src="{{ $e.ThumbnailFile }}" alt="Episode artwork">{{end}}</div>
                            <div>
                                <h3>{{ $e.Title }}</h3>
                                <p>{{ truncate $e.Description 180 }}</p>
                                <div class="row">
                                    <span>{{ episodeDate $e.CreatedAt }}</span><span>·</span><span>{{ $e.Author }}</span>
                                </div>
                            </div>
                        </a>
                    {{ end }}
                {{ end }}
            </div>
        {{ end }}
    </section>

    <footer>
        <a class="brand" href="https://github.com/ofstudio/voxify" target="_blank" rel="noopener">
            <svg viewBox="0 0 32 32" aria-hidden="true">
                <rect x="0" y="0" width="32" height="32" rx="8" fill="#1E40AF"/>
                <svg width="22" height="16" viewBox="0 0 22 16" fill="none" xmlns="http://www.w3.org/2000/svg" x="5"
                     y="8">
                    <path d="M11 0C11.6163 1.34699e-05 12.1162 0.497716 12.1162 1.11133V14.126C12.1162 14.7396 11.6163 15.2373 11 15.2373C10.3837 15.2373 9.88379 14.7396 9.88379 14.126V1.11133C9.88379 0.497708 10.3837 0 11 0ZM7.17383 2.54004C7.614 2.54004 7.97062 2.89478 7.9707 3.33301V11.9043C7.97064 12.3425 7.61402 12.6982 7.17383 12.6982C6.73369 12.6982 6.37701 12.3425 6.37695 11.9043V3.33301C6.37704 2.89481 6.7337 2.5401 7.17383 2.54004ZM14.8262 2.54004C15.2663 2.54008 15.623 2.89481 15.623 3.33301V11.9043C15.623 12.3425 15.2663 12.6982 14.8262 12.6982C14.386 12.6982 14.0294 12.3425 14.0293 11.9043V3.33301C14.0294 2.89478 14.386 2.54004 14.8262 2.54004ZM3.98535 4.44434C4.42558 4.44434 4.78223 4.79998 4.78223 5.23828V10C4.78197 10.4381 4.42542 10.793 3.98535 10.793C3.54533 10.7929 3.18873 10.4381 3.18848 10V5.23828C3.18848 4.80002 3.54517 4.44439 3.98535 4.44434ZM18.0146 4.44434C18.4548 4.44441 18.8115 4.80003 18.8115 5.23828V10C18.8113 10.438 18.4547 10.7929 18.0146 10.793C17.5746 10.793 17.218 10.4381 17.2178 10V5.23828C17.2178 4.79998 17.5744 4.44434 18.0146 4.44434ZM0.796875 5.71387C1.2371 5.71387 1.59375 6.06951 1.59375 6.50781V8.72949C1.59375 9.16779 1.2371 9.52344 0.796875 9.52344C0.356752 9.52332 0 9.16772 0 8.72949V6.50781C0 6.06959 0.356752 5.71399 0.796875 5.71387ZM21.2031 5.71387C21.6433 5.71398 22 6.06958 22 6.50781V8.72949C22 9.16773 21.6433 9.52333 21.2031 9.52344C20.7629 9.52344 20.4062 9.16779 20.4062 8.72949V6.50781C20.4062 6.06951 20.7629 5.71387 21.2031 5.71387Z"
                          fill="white"/>
                </svg>
            </svg>
            <span>Powered by Voxify</span>
        </a>
    </footer>
</div>
</body>
</html>`
