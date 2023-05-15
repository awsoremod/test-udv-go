-- В этой задаче вам предстоит написать запрос к базе данных
-- простейшей системы проведения соревнований по программированию.

-- Схема базы данных содержит четыре таблицы:
-- users
-- contests
-- problems - задачи в системе, каждая задача принадлежит одному контесту
-- submissions - отосланные попытки решения задач, каждая попытка принадлежит
-- одной задаче и одному пользователю


CREATE USER username_1 WITH PASSWORD 'password_1';

CREATE DATABASE first_for_username_1;
GRANT ALL PRIVILEGES ON DATABASE first_for_username_1 TO username_1;
ALTER DATABASE first_for_username_1 OWNER TO username_1;

CREATE DATABASE second_for_username_1;
GRANT ALL PRIVILEGES ON DATABASE second_for_username_1 TO username_1;
ALTER DATABASE second_for_username_1 OWNER TO username_1;

CREATE DATABASE third_for_username_1;
GRANT ALL PRIVILEGES ON DATABASE third_for_username_1 TO username_1;
ALTER DATABASE third_for_username_1 OWNER TO username_1;


CREATE USER username_2 WITH PASSWORD 'password_2';

CREATE DATABASE first_for_username_2;
GRANT ALL PRIVILEGES ON DATABASE first_for_username_2 TO username_2;
ALTER DATABASE first_for_username_2 OWNER TO username_2;

CREATE DATABASE second_for_username_2;
GRANT ALL PRIVILEGES ON DATABASE second_for_username_2 TO username_2;
ALTER DATABASE second_for_username_2 OWNER TO username_2;

CREATE DATABASE third_for_username_2;
GRANT ALL PRIVILEGES ON DATABASE third_for_username_2 TO username_2;
ALTER DATABASE third_for_username_2 OWNER TO username_2;

CREATE DATABASE test1;
CREATE DATABASE test2;

create table users (                                                                      
  id bigint primary key,                                                                           
  name varchar not null                                                                            
);                                                                                                 
                                                                                                   
create table contests (                                                                            
  id bigint primary key,                                                                           
  name varchar not null                                                                            
);                                                                                                 
                                                                                                   
create table problems (                                                                            
  id bigint primary key,                                                                           
  contest_id bigint,                                                                               
  code varchar not null,                                                                           
  constraint fk_problems_contest_id foreign key (contest_id) references contests (id)              
);                                                                                                 
                                                                                                   
create unique index on problems (contest_id, code);                                                
                                                                                                   
create table submissions (                                                                         
  id bigint primary key,                                                                           
  user_id bigint,                                                                                  
  problem_id bigint,                                                                               
  success boolean not null,                                                                        
  submitted_at timestamp not null,                                                                 
  constraint fk_submissions_user_id foreign key (user_id) references users (id),                   
  constraint fk_submissions_problem_id foreign key (problem_id) references problems (id)           
);

insert into users values
(7, 'ABbbAa'),
(21, 'aBB'),
(4, 'abbbbBba'),
(14, 'aA'),
(12, 'ABAaBBBA'),
(19, 'AaAbb'),
(10, 'bBa'),
(11, 'BBABAB'),
(6, 'aAA'),
(9, 'ABb');

insert into contests values
(6, 'B'),
(4, 'A'),
(7, 'A'),
(1, 'a');

insert into problems values
(17, 4, 'A'),
(11, 1, 'A'),
(7, 6, 'A'),
(14, 4, 'B'),
(19, 6, 'B'),
(3, 6, 'C'),
(1, 1, 'B'),
(15, 1, 'C'),
(10, 6, 'D'),
(12, 1, 'D');

insert into submissions values
(87, 4, 11, false, '2023-01-05 11:00:49'),
(129, 14, 7, false, '2023-01-05 11:00:49'),
(194, 11, 15, false, '2023-01-05 11:00:14'),
(151, 11, 7, true, '2023-01-05 11:00:57'),
(50, 11, 15, true, '2023-01-05 11:00:13'),
(137, 4, 17, false, '2023-01-05 11:00:23'),
(138, 11, 7, false, '2023-01-05 11:00:01'),
(195, 4, 12, false, '2023-01-05 11:00:43'),
(155, 14, 12, false, '2023-01-05 11:00:29'),
(31, 11, 12, true, '2023-01-05 11:00:37'),
(77, 6, 11, false, '2023-01-05 11:00:16'),
(79, 10, 15, false, '2023-01-05 11:00:27'),
(189, 6, 12, false, '2023-01-05 11:00:55'),
(24, 6, 12, true, '2023-01-05 11:00:55'),
(121, 19, 1, true, '2023-01-05 11:00:16'),
(13, 10, 11, false, '2023-01-05 11:00:55'),
(4, 4, 3, false, '2023-01-05 11:00:39'),
(110, 19, 3, false, '2023-01-05 11:00:18'),
(46, 4, 1, true, '2023-01-05 11:00:08'),
(44, 9, 15, true, '2023-01-05 11:00:05'),
(16, 21, 3, false, '2023-01-05 11:00:45'),
(169, 9, 1, false, '2023-01-05 11:00:35'),
(157, 4, 17, false, '2023-01-05 11:00:31'),
(70, 6, 19, true, '2023-01-05 11:00:54'),
(191, 6, 1, true, '2023-01-05 11:00:17'),
(101, 14, 3, false, '2023-01-05 11:00:06'),
(39, 4, 17, false, '2023-01-05 11:00:10'),
(21, 11, 12, true, '2023-01-05 11:00:01'),
(29, 19, 12, false, '2023-01-05 11:00:08'),
(196, 9, 3, false, '2023-01-05 11:00:14'),
(178, 14, 15, false, '2023-01-05 11:00:37'),
(34, 12, 12, false, '2023-01-05 11:00:12'),
(173, 10, 17, false, '2023-01-05 11:00:56'),
(152, 10, 15, true, '2023-01-05 11:00:44'),
(95, 19, 17, false, '2023-01-05 11:00:08'),
(143, 10, 17, false, '2023-01-05 11:00:54'),
(166, 10, 11, true, '2023-01-05 11:00:49'),
(22, 19, 15, true, '2023-01-05 11:00:16'),
(135, 4, 11, false, '2023-01-05 11:00:17'),
(75, 4, 7, false, '2023-01-05 11:00:38'),
(54, 12, 12, false, '2023-01-05 11:00:52'),
(109, 10, 15, false, '2023-01-05 11:00:06'),
(170, 10, 3, false, '2023-01-05 11:00:53'),
(100, 12, 7, true, '2023-01-05 11:00:26'),
(73, 9, 19, true, '2023-01-05 11:00:41'),
(128, 7, 1, true, '2023-01-05 11:00:22'),
(187, 4, 12, false, '2023-01-05 11:00:59'),
(10, 4, 1, false, '2023-01-05 11:00:34'),
(146, 12, 3, false, '2023-01-05 11:00:43'),
(145, 11, 15, true, '2023-01-05 11:00:45'),
(11, 4, 12, false, '2023-01-05 11:00:20'),
(172, 10, 1, false, '2023-01-05 11:00:46'),
(115, 9, 17, false, '2023-01-05 11:00:12'),
(47, 6, 7, false, '2023-01-05 11:00:04'),
(132, 7, 1, false, '2023-01-05 11:00:15'),
(167, 10, 3, false, '2023-01-05 11:00:40'),
(127, 4, 15, true, '2023-01-05 11:00:10'),
(12, 14, 12, false, '2023-01-05 11:00:03'),
(182, 10, 11, false, '2023-01-05 11:00:07'),
(180, 6, 12, true, '2023-01-05 11:00:08'),
(90, 11, 10, true, '2023-01-05 11:00:19'),
(99, 6, 15, true, '2023-01-05 11:00:22'),
(40, 10, 17, false, '2023-01-05 11:00:49'),
(161, 9, 12, false, '2023-01-05 11:00:25'),
(102, 10, 7, false, '2023-01-05 11:00:56'),
(80, 14, 19, false, '2023-01-05 11:00:55'),
(179, 14, 14, false, '2023-01-05 11:00:53'),
(28, 14, 3, false, '2023-01-05 11:00:57'),
(175, 4, 19, false, '2023-01-05 11:00:30'),
(19, 9, 1, true, '2023-01-05 11:00:48'),
(165, 9, 15, true, '2023-01-05 11:00:28'),
(55, 19, 17, false, '2023-01-05 11:00:47'),
(176, 14, 15, false, '2023-01-05 11:00:29'),
(1, 4, 3, false, '2023-01-05 11:00:47'),
(200, 14, 3, false, '2023-01-05 11:00:06'),
(61, 10, 17, true, '2023-01-05 11:00:24'),
(130, 14, 12, false, '2023-01-05 11:00:56'),
(60, 4, 19, true, '2023-01-05 11:00:28'),
(134, 10, 12, false, '2023-01-05 11:00:51'),
(74, 7, 17, true, '2023-01-05 11:00:00'),
(136, 4, 7, false, '2023-01-05 11:00:06'),
(64, 11, 17, true, '2023-01-05 11:00:09'),
(186, 4, 3, false, '2023-01-05 11:00:49'),
(63, 19, 12, true, '2023-01-05 11:00:29'),
(201, 14, 15, false, '2023-01-05 11:00:22'),
(8, 6, 15, true, '2023-01-05 11:00:05'),
(147, 10, 11, false, '2023-01-05 11:00:37'),
(6, 11, 1, true, '2023-01-05 11:00:36'),
(181, 9, 17, true, '2023-01-05 11:00:45'),
(197, 4, 3, false, '2023-01-05 11:00:37'),
(88, 11, 3, false, '2023-01-05 11:00:22'),
(84, 14, 15, false, '2023-01-05 11:00:31'),
(15, 7, 3, false, '2023-01-05 11:00:43'),
(43, 9, 15, true, '2023-01-05 11:00:22'),
(48, 4, 12, false, '2023-01-05 11:00:39'),
(190, 12, 7, false, '2023-01-05 11:00:14'),
(81, 9, 15, true, '2023-01-05 11:00:58'),
(119, 4, 15, false, '2023-01-05 11:00:23'),
(168, 9, 12, false, '2023-01-05 11:00:24'),
(131, 9, 11, true, '2023-01-05 11:00:15');

