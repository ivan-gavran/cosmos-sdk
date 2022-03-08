import re
import os
import shutil

def files_to_ignore(parent, all_files):
    trace_pattern = "counterexample[0-9]+.itf.json"
    return [t for t in all_files if (not os.path.isdir(os.path.join(parent, t))) and (re.match(trace_pattern, t) is None)]

apalache_out_relative_path = "../model/_apalache-out"
generated_traces_relative_path = "../generatedTraces"

for test_config_dir in os.listdir(apalache_out_relative_path):
    test_config_dir_path = os.path.join(apalache_out_relative_path, test_config_dir)

    shutil.copytree(test_config_dir_path, 
        generated_traces_relative_path, ignore=files_to_ignore, dirs_exist_ok=True)
                
        

        