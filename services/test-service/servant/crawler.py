from bs4 import BeautifulSoup
import requests

def extract_links_from_url(url:str,depth:int)->list:
    links = []
    if depth == 0:
        return links
    try:
        response = requests.get(url)
        soup = BeautifulSoup(response.text,'html.parser')
        for link in soup.find_all('a'):
            href = link.get('href')
            if href and href.startswith('http'):
                links.append(href)
                links.extend(extract_links_from_url(href,depth-1))
    except:
        pass
    return links 
