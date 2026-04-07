with open('internal/repository/postgres/task_repository.go', 'r', encoding='utf-8') as f:
    lines = f.readlines()
    for i, line in enumerate(lines[110:130], start=111):
        print(f'{i}: {line}', end='')
