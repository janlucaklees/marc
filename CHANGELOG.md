# Changelog

## [0.2.0](https://github.com/janlucaklees/marc/compare/v0.1.0...v0.2.0) (2026-10-07)


### Features

* add -u/--unpaged flag for single-page PDF output ([c2f6e5c](https://github.com/janlucaklees/marc/commit/c2f6e5c00b737c20297cf00035e4de9263d32b09))
* bootstrap starter template/config and skip needless prompts ([73f40ae](https://github.com/janlucaklees/marc/commit/73f40aeac5b4079f66079512bbd9b205d7386a55))
* control template auto-selection via default_template config ([2400017](https://github.com/janlucaklees/marc/commit/24000173e692d9be4efca4ab422fd045fd6293ca))
* default PDF output path from input basename ([5b44168](https://github.com/janlucaklees/marc/commit/5b441689b2a99827b48db30380b730808db5046d))
* detect Chrome/Chromium app bundles on macOS ([ca9a3ef](https://github.com/janlucaklees/marc/commit/ca9a3ef2ff0c12c28b6b715aa4ecef08687f2e47))
* load marc config.toml (pandoc_bin/chromium_bin overrides) ([b6c97e7](https://github.com/janlucaklees/marc/commit/b6c97e712084d779696dc22ddca71906d2e00d2f))
* resolve and interactively prompt for marc templates ([3579528](https://github.com/janlucaklees/marc/commit/3579528ffe8d032cbb373a9d060d458107b37eee))
* resolve pandoc/chromium binaries from config or PATH ([611dca3](https://github.com/janlucaklees/marc/commit/611dca376ceed60e668b57b426497d8ec2b59f5c))
* scaffold marc Go module, port newpage/title preprocessing ([fcdc7de](https://github.com/janlucaklees/marc/commit/fcdc7de65f28310db18dc32566e307a942f227f5))
* wire up marc CLI (render pipeline + flag parsing) ([b638d9a](https://github.com/janlucaklees/marc/commit/b638d9a25df6cc86faa28f3124a0888d8510cf50))


### Bug Fixes

* address Important findings from final branch review (behavior) ([231b33b](https://github.com/janlucaklees/marc/commit/231b33b74a379cb1dc78b1441f7800469d8c5d2f))
* render lists that directly follow a paragraph ([1ed07d0](https://github.com/janlucaklees/marc/commit/1ed07d067ae905554c60b6b3541b7b2beb5730d1))
* use go install for make install instead of manual copy ([993b24d](https://github.com/janlucaklees/marc/commit/993b24dd693975604824644eecce643e1d735dde))
