-- Напишите запрос к базе данных, который возвращает всех пользователей
-- вместе с некоторой дополнительной информацией:
-- 1) Количеством соревнований, в которых он принял участие и решил там
-- хотябы одну задачу
-- 2) Количеством соревнований в которых он принял участие

-- Ожидаемый ответ
-- id | name     | solved_at_least_one_contest_count | take_part_contest_count 
----+----------+-----------------------------------+-------------------------
--  9 | ABb      |                                 3 |                       3 
-- 11 | BBABAB   |                                 3 |                       3 
--  4 | abbbbBba |                                 2 |                       3 
--  7 | ABbbAa   |                                 2 |                       3 
-- 10 | bBa      |                                 2 |                       3 
--  6 | aAA      |                                 2 |                       2 
-- 19 | AaAbb    |                                 1 |                       3 
-- 12 | ABAaBBBA |                                 1 |                       2 
-- 14 | aA       |                                 0 |                       3 
-- 21 | aBB      |                                 0 |                       1 
--(10 rows)

SELECT users.id, users.name,
COUNT
  (
    DISTINCT CASE WHEN submissions.success = true
    THEN problems.contest_id END
  )
  AS solved_at_least_one_contest_count,

COUNT(DISTINCT problems.contest_id) AS take_part_contest_count

FROM users
LEFT JOIN submissions ON users.id = submissions.user_id
LEFT JOIN problems ON submissions.problem_id = problems.id
GROUP BY users.id
ORDER BY solved_at_least_one_contest_count DESC, take_part_contest_count DESC, users.id
