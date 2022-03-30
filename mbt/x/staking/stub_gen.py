def stub(x):
    return f"""func TestTracesP{x}(t *testing.T) {{
	ExecuteTraces(t, loadTraces("traces/P{x}.json"))
}}"""

for i in range(13, 37):
    print(stub(i))