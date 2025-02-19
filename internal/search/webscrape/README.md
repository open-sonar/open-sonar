## Setup Web Scraper

1. **Init:**
   ```sh
   make init
   ```

2. **Build scraper:**
   ```sh
   make build
   ```

3. **Search query example (defaults to first 3 pages):**
   ```sh
   make run QUERY="Bulbasaur"
   ```

4. **Specific page no. search query example:**
   ```sh
   make run QUERY="Bulbasaur" PAGES=1
   ```

5. **Clean up built binary:**
   ```sh
   make clean
   ```

6. **Update dependencies:**
   ```sh
   make deps
   ```

