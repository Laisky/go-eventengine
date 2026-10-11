The candidate preserves the existing module identity and major application API. The patched golang.org/x/net v0.60.0 dependency requires Go 1.26.0 or newer; qualification uses Go1.26.9. Module verification, vet, build, full race tests and source vulnerability checks pass. No security finding is suppressed.

Your code is affected by 0 vulnerabilities.
This scan also found 2 vulnerabilities in packages you import and 1
vulnerability in modules you require, but your code doesn't appear to call these
vulnerabilities.
Reachability results do not guarantee that every unused module or downstream consumer is vulnerability-free.


CodeQL extraction and analysis are preserved, with official v4 action support and Go1.26.9 setup. Existing main/develop pull-request, push and weekly triggers and security-events write permission remain. Hosted CodeQL qualification must pass on the published head before merge.
