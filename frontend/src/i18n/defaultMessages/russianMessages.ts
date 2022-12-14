import { TranslationMessages } from 'ra-core';

const russianMessages: TranslationMessages = {
    ra: {
        action: {
            add_filter: 'Добавить фильтр',
            add: 'Добавить',
            back: 'Назад',
            bulk_actions: '1 элемент выбран |||| %{smart_count} выбрано элементов',
            cancel: 'Отмена',
            clear_array_input: 'Очистить список',
            clear_input_value: 'Очистить значение',
            clone: 'Скопировать',
            confirm: 'Подтвердить',
            create: 'Создать',
            create_item: 'Создать %{item}',
            delete: 'Удалить',
            edit: 'Изменить',
            export: 'Экспорт',
            list: 'Список',
            refresh: 'Обновить',
            remove_filter: 'Убрать фильтр',
            remove_all_filters: 'Убрать все фильтры',
            remove: 'Переместить',
            save: 'Сохранить',
            search: 'Поиск',
            select_all: 'Выбрать все',
            select_row: 'Выбрать строку',
            select_columns: 'Выбрать поля',
            show: 'Показать',
            sort: 'Сортировка',
            undo: 'Отменить',
            unselect: 'Отменить выделение',
            expand: 'Раскрыть',
            close: 'Закрыть',
            open_menu: 'Открыть меню',
            close_menu: 'Закрыть меню',
            update: 'Обновить',
            move_up: 'Переместить вверх',
            move_down: 'Переместить вниз',
            open: 'Открыть',
            toggle_theme: "Переключить тему"
        },
        boolean: {
            true: 'Да',
            false: 'Нет',
            null: ' ',
        },
        page: {
            create: 'Создать %{name}',
            dashboard: 'Панель управления',
            edit: '%{name} #%{id}',
            error: 'Ошибка выполнения операции',
            list: '%{name}',
            loading: 'Загрузка',
            not_found: 'Не найдено',
            show: '%{name} #%{id}',
            empty: 'Нет %{name}',
            invite: 'Вы хотите добавить еще одну?',
        },
        input: {
            file: {
                upload_several: "Перетащите файлы сюда или нажмите для выбора.",
                upload_single: "Перетащите файл сюда или нажмите для выбора."
            },
            image: {
                upload_several: "Перетащите изображения сюда или нажмите для выбора.",
                upload_single: "Перетащите изображение сюда или нажмите для выбора."
            },
            references: {
                all_missing: "Связанных данных не найдено",
                many_missing:
                    "Некоторые из связанных данных не доступны",
                single_missing:
                    "Связанный объект не доступен"
            },
            password: {
                toggle_visible: 'Скрыть пароль',
                toggle_hidden: 'Показать пароль',
            },
        },
        message: {
            about: "Справка",
            are_you_sure: "Вы уверены?",
            clear_array_input: 'Вы уверены, что хотите очистить вест список?',
            bulk_delete_content:
                "Вы уверены, что хотите удалить %{name}? |||| Вы уверены, что хотите удалить объекты, кол-вом %{smart_count} ? |||| Вы уверены, что хотите удалить объекты, кол-вом %{smart_count} ?",
            bulk_delete_title: "Удалить %{name} |||| Удалить %{smart_count} %{name} |||| Удалить %{smart_count} %{name}",
            delete_content: "Вы уверены что хотите удалить этот объект",
            delete_title: "Удалить %{name} #%{id}",
            details: "Описание",
            error: "В процессе запроса возникла ошибка, и он не может быть завершен",
            invalid_form: "Форма заполнена неверно, проверьте, пожалуйста, ошибки",
            loading: "Идет загрузка, пожалуйста, подождите...",
            no: "Нет",
            not_found: "Ошибка URL или вы следуете по неверной ссылке",
            yes: "Да",
            unsaved_changes:
                "Некоторые из ваших изменений не были сохранены. Вы уверены, что хотите их игнорировать?",
            bulk_update_content:
                'Вы уверены что хотите обновить этот %{name}? |||| Вы уверены что хотите обновить эти %{smart_count} элементы?',
            bulk_update_title:
                'Обновление %{name} |||| Обновление %{smart_count} %{name}',
        },
        navigation: {
            no_results: "Результатов не найдено",
            no_more_results:
                "Страница %{page} выходит за пределы нумерации, попробуйте предыдущую",
            page_out_of_boundaries: "Страница %{page} вне границ",
            page_out_from_end: "Невозможно переместиться дальше последней страницы",
            page_out_from_begin: "Номер страницы не может быть меньше 1",
            page_range_info: "%{offsetBegin}-%{offsetEnd} из %{total}",
            page_rows_per_page: "Строк на странице:",
            next: "Следующая",
            prev: "Предыдущая",
            skip_nav: "Перейти к содержанию",
            partial_page_range_info:
                '%{offsetBegin}-%{offsetEnd} из более чем %{offsetEnd}',
            current_page: 'Страница %{page}',
            page: 'Перейти к странице %{page}',
            first: 'Первая',
            last: 'Последняя',
            previous: 'Предыдущая',
        },
        sort: {
            sort_by: 'Сортировать по %{field} %{order}',
            ASC: 'возрастанию',
            DESC: 'убыванию',
        },
        auth: {
            auth_check_error: "Пожалуйста, авторизуйтесь для продолжения работы",
            user_menu: "Профиль",
            username: "Имя пользователя",
            password: "Пароль",
            sign_in: "Войти",
            sign_in_error: "Ошибка аутентификации, попробуйте снова",
            logout: "Выйти"
        },
        notification: {
            updated: "Элемент обновлен |||| %{smart_count} обновлено |||| %{smart_count} обновлено",
            created: "Элемент создан",
            deleted: "Элемент удален |||| %{smart_count} удалено |||| %{smart_count} удалено",
            bad_item: "Элемент не валиден",
            item_doesnt_exist: "Элемент не существует",
            http_error: "Ошибка сервера",
            data_provider_error: "Ошибка dataProvider, проверьте консоль",
            i18n_error: "Не удалось загрузить перевод для указанного языка",
            canceled: "Операция отменена",
            logged_out: "Ваша сессия завершена, попробуйте переподключиться/войти снова",
            not_authorized: "У Вас нет доступа к данному ресурсу.",
        },
        validation: {
            required: "Обязательно для заполнения",
            minLength: "Минимальное кол-во символов %{min}",
            maxLength: "Максимальное кол-во символов %{max}",
            minValue: "Минимальное значение %{min}",
            maxValue: "Значение может быть %{max} или меньше",
            number: "Должно быть цифрой",
            email: "Некорректный email",
            oneOf: "Должно быть одним из: %{options}",
            regex: "Должно быть в формате (regexp): %{pattern}"
        },
        saved_queries: {
            label: 'Сохраненные запросы',
            query_name: 'Имя запроса',
            new_label: 'Сохранить текущий запрос...',
            new_dialog_title: 'Сохранить текущий запрос как...',
            remove_label: 'Удалить запрос',
            remove_label_with_name: 'Удалить запрос "%{name}"',
            remove_dialog_title: 'Удаление сохраненных запросов?',
            remove_message:
                'Вы уверены, что хотите удалить этот элемент из списка сохраненных запросов?',
            help: 'Отфильтровать список и сохраните этот запрос',
        },
    }
};

export default russianMessages;