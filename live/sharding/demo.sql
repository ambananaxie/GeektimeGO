select sum(amount) from `order` group by user_id

select sum(amount) as amount from `order` group by user_id, product_id

select sum(amount) from `order` group by user_id order by id;


delete from user_tab where id IN (1, 2, 3);

delete from user_tab_1 where id IN (1, 2, 3);
delete from user_tab_2 where id IN (1, 2, 3);
delete from user_tab_3 where id IN (1, 2, 3);

Begin;

delete from user_tab where id IN (1, 2, 3); -- user_tab_1, user_tab_2, user_tab_3

delete from user_tab where id IN (2, 4, 7); -- user_tab_2, user_tab_4, user_tab_7
commit ;