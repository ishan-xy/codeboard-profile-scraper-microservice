import asyncio
import aiohttp
import json
import nats
from nats.errors import TimeoutError
import os
import time

servers = os.environ.get("NATS_URL", "nats://localhost:4222").split(",")

async def fetch(session, username, query, limit=20):
    url = "https://leetcode.com/graphql/"
    payload = {
        "query": query,
        "variables": {
            "username": username,
            "limit": limit
        },
        "operationName": "recentAcSubmissions"
    }
    
    async with session.post(url, json=payload) as response:
        if response.status != 200:
            response.raise_for_status()
        data = await response.json()
        return {username: data.get("data", {}).get("recentAcSubmissionList", [])}

async def fetch_all(session, usernames, query):
    tasks = [fetch(session, username, query) for username in usernames]
    results = await asyncio.gather(*tasks, return_exceptions=True)
    combined_results = {}
    for result in results:
        if isinstance(result, Exception):
            print(f"Error fetching data: {result}")
            continue
        combined_results.update(result)
    return combined_results

async def handle_request(msg):
    try:
        start_time = time.time()  # Start timing the process
        
        data = json.loads(msg.data.decode())
        usernames = data.get("usernames", [])
        query = data.get("query", "")
        
        print(f"Processing usernames: {usernames}")
        
        async with aiohttp.ClientSession() as session:
            combined_results = await fetch_all(session, usernames, query)
            
            # Calculate processing time
            process_time = time.time() - start_time
            print(f"Processing time: {process_time:.2f} seconds")
            
            # Add processing time to the response
            response_data = {
                "results": combined_results,
                "process_time": process_time
            }
            
            # Send the response back to the publisher
            response = json.dumps(response_data).encode()
            await msg.respond(response)
            print("Response sent back to publisher")
    
    except Exception as e:
        print(f"Error processing request: {e}")

async def main():
    nc = await nats.connect(servers=servers)
    sub = await nc.subscribe("usernames", cb=handle_request)
    
    print("Subscriber is running and waiting for requests...")
    
    # Keep the subscriber running
    while True:
        await asyncio.sleep(1)

if __name__ == '__main__':
    asyncio.run(main())