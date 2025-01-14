import argparse
import docker
import gzip
import yaml
from nested_lookup import nested_lookup


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Save Docker image")

    parser.add_argument("--manifest",
                        required=True,
                        help="Path to the manifest yaml")
    arguments = parser.parse_args()


    with open(arguments.manifest, 'r') as file:
        template = file.read()
    
    images=[]
    parts = template.split('---')
    for p in parts:
        y = yaml.safe_load(p)
        matches = nested_lookup("image", y)
        if (len(matches)):
                images += matches
    
    client = docker.from_env()
    for image_name in set(images):
        file_name = (image_name.split(':')[0].replace("/", "-"))
        f = gzip.open( file_name + '.tar.gz', 'wb')
        try:
            image = client.images.get(image_name)
            if image.id:
              print ("docker image \"" + image_name + "\" already exists.")
        except docker.errors.ImageNotFound:
            print ("docker pull " + image_name + " ...")
            image = client.images.pull(image_name)
        image_tar = image.save(named=True)
        f.writelines(image_tar)
        f.close()
