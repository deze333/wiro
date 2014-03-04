Web Intelligent Resource Objects
================================
Resource repository for web applications
----------------------------------------

Wiro is a compact repository for flexible content retrieval based on request key:
* domain
* language
* A/B testing weight

All resources are memcached for max performance.

Repository files are watched for changes (via Linux inotify) and memory cache updated upon changes.
