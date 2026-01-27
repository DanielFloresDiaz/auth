#!/usr/bin/env python3
"""Generate Kubernetes Secret YAML for GHCR authentication.

Example usage:
    python generate_ghcr_secret.py -u myuser -t mytoken -n ns1 ns2 ns3 --apply

This script generates the necessary base64 encoded authentication string,
the Docker config JSON, and creates Kubernetes Secret YAML files
for each specified namespace for pulling images from GitHub Container Registry (GHCR).
It uses the ghcr_secret.yaml template file for the YAML structure.
Optionally, apply the secrets directly to the cluster using --apply.
"""

import argparse
import base64
import json
import os
import subprocess


def main():
    # Load the template
    script_dir = os.path.dirname(__file__)
    template_path = os.path.join(script_dir, 'ghcr_secret.yaml')
    with open(template_path) as f:
        template = f.read()

    parser = argparse.ArgumentParser(description='Generate Kubernetes Secret YAML for GHCR authentication')
    parser.add_argument('-u', '--user', required=True, help='GitHub username')
    parser.add_argument('-t', '--token', required=True, help='GitHub token')
    parser.add_argument('-n', '--namespace', required=True, nargs='+', help='Kubernetes namespace(s)')
    parser.add_argument('-o', '--output', default='ghcr_secret', help='Output file prefix (default: ghcr_secret)')
    parser.add_argument(
        '--secret-name', default='ghcr-secret', help='Name of the Kubernetes secret (default: ghcr-secret)'
    )
    parser.add_argument('--apply', action='store_true', help='Apply the secrets to Kubernetes using kubectl')
    parser.add_argument('-v', '--verbose', action='store_true', help='Enable verbose output')

    args = parser.parse_args()

    username = args.user
    token = args.token

    # Step 1: Create base64 encoded username:token
    auth_string = f'{username}:{token}'
    auth_b64 = base64.b64encode(auth_string.encode('utf-8')).decode('utf-8')
    if args.verbose:
        print(f'Base64 encoded auth: {auth_b64}')

    # Step 2: Create the dict
    config = {'auths': {'ghcr.io': {'auth': auth_b64}}}

    # Step 3: Base64 encode the dict
    json_str = json.dumps(config, indent=2)
    final_b64 = base64.b64encode(json_str.encode('utf-8')).decode('utf-8')
    if args.verbose:
        print(f'Base64 encoded docker config: {final_b64}')

    # Step 4: Generate the Kubernetes Secret YAML for each namespace
    for ns in args.namespace:
        yaml_content = (
            template.replace('{NAMESPACE}', ns)
            .replace('{BASE64_ENCODED_DOCKER_CONFIG_JSON}', final_b64)
            .replace('{SECRET_NAME}', args.secret_name)
        )

        output_file = f'{args.output}_{ns}.yaml'
        with open(output_file, 'w') as f:
            f.write(yaml_content)

        print(f'Kubernetes Secret YAML for namespace {ns} written to {output_file}')

        if args.apply:
            subprocess.run(['kubectl', 'apply', '-f', output_file], check=True)
            print(f'Applied secret for namespace {ns} to Kubernetes')


if __name__ == '__main__':
    main()
