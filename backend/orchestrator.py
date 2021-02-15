import json
import os
import shutil
import tarfile
import time
from email.mime.text import MIMEText
from smtplib import SMTP_SSL as SMTP

from paramiko import SSHClient, AutoAddPolicy
from paramiko.ssh_exception import NoValidConnectionsError
from pymongo import MongoClient
from pystalk import BeanstalkClient
from scp import SCPClient

import utils
from config import configuration

db_client = MongoClient("mongodb://localhost:27017")
db = db_client["cdn"]

queue = BeanstalkClient("localhost", 11300)

ssh = SSHClient()
ssh.load_system_host_keys()
ssh.set_missing_host_key_policy(AutoAddPolicy())


def normalize(string: str) -> str:
    return string.rstrip(".").upper().replace("-", "_").replace(".", "_").replace("@", "_").replace(".", "_")


def run_ssh_command(command):
    print("    - running " + command, end="", flush=True)
    stdin, stdout, stderr = ssh.exec_command(command)
    for line in stdout:
        print(" - " + line.strip('\n'))
    for line in stderr:
        print(" - ERR " + line.strip('\n'))


print("Starting main loop")

while True:
    for job in queue.reserve_iter():
        content = json.loads(job.job_data)

        operation = content["operation"]
        args = content["args"]

        # if operation == "refresh_single_zone":
        #     print("refreshing " + args["zone"])
        #
        #     zone = db["zones"].find_one({"zone": args["zone"]})
        #
        #     if not args or not args.get("node"):
        #         _query = {}
        #     else:
        #         if args["node"] == "all":
        #             _query = {}
        #         else:
        #             _query = {"name": args["node"]}
        #
        #     # Loop over the nodes and send the updated zone file to each one, then reload the configuration
        #     for node in db["nodes"].find(_query):
        #         print("... now updating " + node["name"] + " " + node["management_ip"] + " " + node["location"])
        #
        #         zone_file = utils.render_zone(zone, node)
        #
        #         with open("/tmp/db." + zone["zone"], "w") as zone_file_writer:
        #             zone_file_writer.write(zone_file)
        #
        #         print("    - sending updated zone file")
        #         try:
        #             ssh.connect(node["management_ip"], username="root", port=34553, key_filename=configuration["ssh-key"])
        #         except (TimeoutError, NoValidConnectionsError):
        #             error = "- ERROR: " + node["name"] + " timed out."
        #             print(error)
        #         else:
        #             with SCPClient(ssh.get_transport()) as scp:
        #                 scp.put("/tmp/db." + zone["zone"], "/etc/bind/db." + zone["zone"])
        #
        #             run_ssh_command("rndc reload")
        #             ssh.close()
        #
        #     print("finished refresh_single_zone")

        if operation == "refresh_all_zones":
            print("refreshing local zones file")

            try:
                shutil.rmtree("/tmp/packetframedeploy/")
            except FileNotFoundError:
                pass

            os.mkdir("/tmp/packetframedeploy/")

            for node in db["nodes"].find():
                print(f"Build archive for {node['name']}")
                os.mkdir("/tmp/packetframedeploy/" + node["name"])

                zone_registry = ""
                for zone in db["zones"].find():
                    zone_registry += utils.render_local(zone["zone"])

                    zone_file = utils.render_zone(zone, node)

                    with open(f"/tmp/packetframedeploy/{node['name']}/db.{zone['zone']}", "w") as zone_file_writer:
                        zone_file_writer.write(zone_file)

                with open(f"/tmp/packetframedeploy/{node['name']}/named.conf.local", "w") as named_file:
                    named_file.write(zone_registry)

                with tarfile.open(f"/tmp/packetframedeploy/{node['name']}.tar.gz", "w:gz") as tar:
                    output_file = f"/tmp/packetframedeploy/{node['name']}/"
                    tar.add(output_file, arcname=os.path.basename(output_file))

                try:
                    ssh.connect(node["management_ip"], username="root", port=34553, key_filename=configuration["ssh-key"])
                except (TimeoutError, NoValidConnectionsError):
                    print(node["name"] + " unable to connect.")
                else:
                    with SCPClient(ssh.get_transport()) as scp:
                        scp.put(f"/tmp/packetframedeploy/{node['name']}.tar.gz", "/etc/bind/deploy.tar.gz")

                    stdin, stdout, stderr = ssh.exec_command("tar -xvzf /etc/bind/deploy.tar.gz -C /etc/bind/ ; rm /etc/bind/deploy.tar.gz ; bash /root/cleanup-zones.sh ; rndc reload")
                    for line in stdout:
                        print(line.strip())
                    if line in stderr:
                        print(line.strip())

            # Clean up the tmp files
            shutil.rmtree("/tmp/packetframedeploy/")

            print("finished refresh_all_zones task")

        elif operation == "delete_zone":
            print("deleting " + args["zone"])

            for node in db["nodes"].find():
                print("... now updating " + node["name"] + " " + node["management_ip"] + " " + node["location"])

                try:
                    ssh.connect(node["management_ip"], username="root", port=34553, key_filename=configuration["ssh-key"])
                except (TimeoutError, NoValidConnectionsError):
                    error = "- ERROR: " + node["name"] + " timed out."
                    print(error)
                else:
                    run_ssh_command("rm /etc/bind/db." + args["zone"])
                    ssh.close()

            print("finished delete_zone")

        elif operation == "node_power":
            print("setting node " + args["ip"] + " to " + args["state"])

            try:
                ssh.connect(args["ip"], username="root", port=34553, key_filename=configuration["ssh-key"])
            except (TimeoutError, NoValidConnectionsError):
                error = "- ERROR: " + args["ip"] + " timed out."
                print(error)
            else:
                if args["state"] == "on":
                    run_ssh_command("systemctl start bird")
                elif args["state"] == "off":
                    run_ssh_command("birdc down")
                ssh.close()

            print("finished node_power")

        elif operation == "refresh_cache":
            backends = {}
            domains = {}
            acls = {}

            for zone in db["zones"].find():
                for record in zone["records"]:
                    if record.get("proxied"):
                        domain = record["label"].rstrip(".")
                        safe_name = "BACKEND_" + normalize(domain)
                        backends[safe_name] = record["value"]
                        domains[domain] = safe_name

                        acl_name = "ACL_" + normalize(zone["zone"])

                        if acl_name not in acls:
                            acls[acl_name] = []
                            acl = acls[acl_name]

                            for user in zone["users"]:
                                user_doc = db["users"].find_one({"username": user})
                                if "acl" in user_doc:
                                    for address in user_doc["acl"]:
                                        acl.append(address)

            # Render and write the default.vcl tmp file
            with open("/tmp/default.vcl", "w") as vcl_file:
                vcl_file.write(utils.render_vcl(backends, domains, acls, normalize))

            # Deploy the vcl file and reload
            for node in db["nodes"].find({"http": True}):
                print("... now updating " + node["name"] + " " + node["management_ip"] + " " + node["location"])
                print("    - sending updated vcl/caddy configs")

                # Render and write the Caddyfile tmp file
                with open("/tmp/Caddyfile", "w") as caddy_file:
                    caddy_file.write(utils.render_caddy(domains, node))

                try:
                    ssh.connect(node["management_ip"], username="root", port=34553, key_filename=configuration["ssh-key"])
                except (TimeoutError, NoValidConnectionsError):
                    error = "- ERROR: " + node["name"] + " timed out."
                    print(error)
                else:
                    with SCPClient(ssh.get_transport()) as scp:
                        scp.put("/tmp/default.vcl", "/etc/varnish/default.vcl")
                    with SCPClient(ssh.get_transport()) as scp:
                        scp.put("/tmp/Caddyfile", "/etc/caddy/Caddyfile")

                    run_ssh_command("systemctl reload varnish")
                    run_ssh_command("caddy reload -config /etc/caddy/Caddyfile")
                    ssh.close()

            print("finished refresh_cache")

        elif operation == "send_email":
            recipients = args["recipients"]
            message = args["body"]
            subject = args["subject"]
            print("Sending email to", recipients)

            msg = MIMEText(message, "plain")
            msg["Subject"] = subject
            msg["From"] = configuration["email"]["username"]

            server = SMTP(configuration["email"]["server"])
            server.login(configuration["email"]["username"], configuration["email"]["password"])
            server.sendmail(configuration["email"]["username"], recipients + ["info@packetframe.com"], msg.as_string())
            server.quit()
            print("Sent")

        elif operation == "update_collector":
            try:
                ssh.connect(configuration["collector"]["host"], username="root", port=34553, key_filename=configuration["ssh-key"])
            except (TimeoutError, NoValidConnectionsError):
                error = "- Collector timed out."
                print(error)
            else:
                with SCPClient(ssh.get_transport()) as scp:
                    scp.put("/tmp/collector_bird.conf", "/etc/bird/bird.conf")

                run_ssh_command("birdc conf")
                ssh.close()

        elif operation == "update_monitoring":
            print("Deploying monitoring config...")
            # noinspection StandardShellInjection
            os.system("scp -P 34553 -i /home/nate/ssh-key prometheus.yml " + configuration["monitoring_host"] + ":/etc/prometheus/prometheus.yml")
            # noinspection StandardShellInjection
            os.system("ssh -p 34553 " + configuration["monitoring_host"] + " -i /home/nate/ssh-key 'curl -X POST http://localhost:9090/-/reload'")
            print("Finished deploying monitoring config")

        elif operation == "healthcheck":
            print("Running healthcheck")
            for node in db["nodes"].find():
                print(f"running healthcheck {node['location']}")

                try:
                    ssh.connect(node["management_ip"], username="root", port=34553, key_filename=configuration["ssh-key"])
                except (TimeoutError, NoValidConnectionsError):
                    error = "- ERROR: " + node["name"] + " timed out."
                    print(error)
                else:
                    _, state_stdout, _ = ssh.exec_command("systemctl is-active bind9")
                    state_stdout = state_stdout.readlines()
                    if "".join(state_stdout).strip("\n") != "active":
                        print(node["name"] + " ERR")
                    else:
                        print(node["name"] + " OK")

            else:
                print("ERROR: This task isn't recognized")

        queue.delete_job(job.job_id)

    time.sleep(0.5)
