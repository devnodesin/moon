# Automated API test runner and result saver

import requests
import json
from datetime import datetime
import argparse
import os



def build_curl_command(url, method, headers, data):
	"""
	Builds a formatted curl command with proper line continuations and indentation.
	"""
	curl_lines = [f'curl -s -X {method} "{url}"']
	if headers:
		for k, v in headers.items():
			curl_lines.append(f'    -H "{k}: {v}"')
	if data is not None:
		if isinstance(data, dict):
			# Format JSON with proper indentation (4 spaces base, 2 spaces for structure)
			pretty_data = json.dumps(data, indent=2)
			# Indent each line of the JSON by 4 spaces
			indented_lines = []
			for line in pretty_data.split('\n'):
				indented_lines.append('      ' + line)
			indented_data = '\n'.join(indented_lines)
			curl_lines.append(f"    -d '\n{indented_data}\n    '")
		else:
			curl_lines.append(f"    -d '{data}'")
	# Add trailing backslash for all but last line
	for i in range(len(curl_lines)-1):
		curl_lines[i] += ' \\'
	curl_cmd = "\n".join(curl_lines) + " | jq ."
	return curl_cmd

def run_test(base_url, prefix, test):
	"""
	Executes a single API request using the given base URL, prefix, and test definition.
	Returns the curl command, response status, and response body (as string).
	"""
	method = test.get("cmd", "GET").upper()
	endpoint = test.get("endpoint", "/")
	url = f"{base_url}{prefix}{endpoint}"
	headers = test.get("headers", {})
	data = test.get("data")
	req_kwargs = {}
	if headers:
		req_kwargs["headers"] = headers
	if data is not None:
		if isinstance(data, dict):
			req_kwargs["json"] = data
		else:
			req_kwargs["data"] = data
	
	curl_cmd = build_curl_command(url, method, headers, data)
	
	try:
		resp = requests.request(method, url, **req_kwargs)
		status = f"{resp.status_code} {resp.reason}"
		try:
			body = resp.json()
			body_str = json.dumps(body, indent=2)
		except Exception:
			body_str = resp.text
	except Exception as e:
		status = "ERROR"
		body_str = str(e)
	return curl_cmd, status, body_str


def parse_args():
	"""
	Parses command-line arguments for output directory.
	"""
	parser = argparse.ArgumentParser(description="Automated API test runner")
	parser.add_argument('-o', '--outdir', default='./out', help='Output directory for result files (default: ./out)')
	parser.add_argument('-i', '--input', default=None, help='Test JSON file to run (default: all in tests dir)')
	parser.add_argument('-t', '--testdir', default='./tests', help='Directory containing test JSON files (default: ./tests)')
	return parser.parse_args()

def setup_outdir(outdir):
	"""
	Ensures the output directory exists (creates if missing).
	"""
	os.makedirs(outdir, exist_ok=True)

def format_markdown_result(curl_cmd, status, body, test_name=None):
	"""
	Formats a single API test result as a Markdown snippet for output, with heading if test_name is given.
	"""
	heading = f"**{test_name}**\n\n" if test_name else ""
	return [
		heading + f"```bash\n{curl_cmd}\n```",
		f"\n***Response ({status}):***\n",
		f"```json\n{body}\n```\n"
	]

def run_all_tests(tests, outdir, access_token=None, outfilename=None):
	"""
	Runs all API tests and writes Markdown output to the output file.
	Returns status (success/failure) and output file path.
	"""
	results_md = []
	docURL = tests["docURL"]
	serverURL = tests["serverURL"]
	prefix = tests.get("prefix", "")
	all_ok = True
	for test in tests["tests"]:
		curl_cmd, status, body = run_test(serverURL, prefix, test)
		# Replace actual server URL with doc URL for display
		curl_cmd_doc = curl_cmd.replace(serverURL, docURL)
		# Replace actual access token with placeholder for documentation
		if access_token:
			curl_cmd_doc = curl_cmd_doc.replace(access_token, "$ACCESS_TOKEN")
		test_name = test.get("name")
		results_md.extend(format_markdown_result(curl_cmd_doc, status, body, test_name))
		if not status.startswith("2"):
			all_ok = False
	markdown = "\n".join(results_md)
	if outfilename:
		with open(outfilename, "w", encoding="utf-8") as f:
			f.write(markdown)
	return ("success" if all_ok else "failure", outfilename, markdown)

def main():
	"""
	Entry point: parses arguments, ensures output directory, and runs all tests. Prints only status and markdown output.
	"""
	args = parse_args()
	setup_outdir(args.outdir)
	test_files = []
	if args.input:
		test_files = [args.input]
	else:
		# Find all .json files in testdir
		test_files = [os.path.join(args.testdir, f) for f in os.listdir(args.testdir) if f.endswith('.json')]
	for test_file in test_files:
		with open(test_file, 'r', encoding='utf-8') as f:
			tests = json.load(f)

		# Perform health check if specified
		health_endpoint = tests.get("health", "/health")
		health_url = f"{tests['serverURL']}{tests.get('prefix', '')}{health_endpoint}"
		try:
			health_resp = requests.get(health_url, timeout=5)
			if health_resp.status_code != 200:
				print(f"Skipping {test_file} [server unhealthy: {health_resp.status_code}]")
				continue
		except Exception as e:
			print(f"Skipping {test_file} [server unreachable: {e}]")
			continue

		# Check if any test uses Authorization header with $ACCESS_TOKEN
		need_token = any(
			'headers' in t and 'Authorization' in t['headers'] and '$ACCESS_TOKEN' in t['headers']['Authorization']
			for t in tests.get('tests', [])
		)
		access_token = None
		if need_token:
			# Perform login
			login_url = f"{tests['serverURL']}/auth:login"
			login_data = {
				"username": tests.get("username", "admin"),
				"password": tests.get("password", "moonadmin12#")
			}
			try:
				resp = requests.post(login_url, json=login_data, headers={"Content-Type": "application/json"})
				resp.raise_for_status()
				token_json = resp.json()
				access_token = token_json.get("access_token")
			except Exception as e:
				print(f"Login failed: {e}")
				access_token = None
		# Replace $ACCESS_TOKEN in test headers
		for t in tests.get('tests', []):
			if 'headers' in t and 'Authorization' in t['headers'] and '$ACCESS_TOKEN' in t['headers']['Authorization']:
				if access_token:
					t['headers']['Authorization'] = t['headers']['Authorization'].replace('$ACCESS_TOKEN', access_token)
				else:
					t['headers']['Authorization'] = t['headers']['Authorization'].replace('$ACCESS_TOKEN', '')

		# Output file: out/<basename>.md
		base = os.path.splitext(os.path.basename(test_file))[0]
		outfilename = os.path.join(args.outdir, f"{base}.md")
		status, outfile, markdown = run_all_tests(tests, args.outdir, access_token, outfilename)
		print("\n==============================================")
		print(f"Executed {test_file} [{status}]")
		print("==============================================\n")
		print(markdown)

if __name__ == "__main__":
	main()
