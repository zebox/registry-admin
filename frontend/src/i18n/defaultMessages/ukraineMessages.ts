import { TranslationMessages } from 'ra-core';

const ukraineMessages: TranslationMessages = {
    ra: {
        action: {
            add_filter: 'Додати фільтр',
            add: 'Додати',
            back: 'Назад',
            bulk_actions: '1 елемент вибрано |||| %{smart_count} вибрано елементів',
            cancel: 'Скасувати',
            clear_array_input: 'Очистити список',
            clear_input_value: 'Очистити значення',
            clone: 'Скопіювати',
            confirm: 'Підтвердити',
            create: 'Створити',
            create_item: 'Створити %{item}',
            delete: 'Видалити',
            edit: 'Змінити',
            export: 'Експорт',
            list: 'Список',
            refresh: 'Оновити',
            remove_filter: 'Прибрати фільтр',
            remove_all_filters: 'Прибрати всі фільтри',
            remove: 'Перемістити',
            save: 'Зберегти',
            search: 'Пошук',
            select_all: 'Вибрати все',
            select_row: 'Вибрати рядок',
            select_columns: 'Вибрати поля',
            show: 'Показати',
            sort: 'Сортування',
            undo: 'Скасувати',
            unselect: 'Скасувати виділення',
            expand: 'Розгорнути',
            close: 'Закрити',
            open_menu: 'Відкрити меню',
            close_menu: 'Закрити меню',
            update: 'Оновити',
            move_up: 'Перемістити вгору',
            move_down: 'Перемістити вниз',
            open: 'Відкрити',
            toggle_theme: "Переключити тему"
        },
        boolean: {
            true: 'Так',
            false: 'Ні',
            null: ' ',
        },
        page: {
            create: 'Створити %{name}',
            dashboard: 'Панель керування',
            edit: '%{name} #%{id}',
            error: 'Помилка виконання операції',
            list: '%{name}',
            loading: 'Завантаження',
            not_found: 'Не знайдено',
            show: '%{name} #%{id}',
            empty: 'Немає %{name}',
            invite: 'Бажаєте додати ще один?',
        },
        input: {
            file: {
                upload_several: "Перетягніть файли сюди або натисніть для вибору.",
                upload_single: "Перетягніть файл сюди або натисніть для вибору."
            },
            image: {
                upload_several: "Перетягніть зображення сюди або натисніть для вибору.",
                upload_single: "Перетягніть зображення сюди або натисніть для вибору."
            },
            references: {
                all_missing: "Пов'язаних даних не знайдено",
                many_missing:
                    "Деякі з пов'язаних даних недоступні",
                single_missing:
                    "Пов'язаний об'єкт недоступний"
            },
            password: {
                toggle_visible: 'Приховати пароль',
                toggle_hidden: 'Показати пароль',
            },
        },
        message: {
            about: "Довідка",
            are_you_sure: "Ви впевнені?",
            auth_error: "Сталася помилка при перевірці токена автентифікації",
            clear_array_input: 'Ви впевнені, що хочете очистити весь список?',
            bulk_delete_content:
                "Ви впевнені, що хочете видалити %{name}? |||| Ви впевнені, що хочете видалити об'єкти, кількістю %{smart_count}? |||| Ви впевнені, що хочете видалити об'єкти, кількістю %{smart_count}?",
            bulk_delete_title: "Видалити %{name} |||| Видалити %{smart_count} %{name} |||| Видалити %{smart_count} %{name}",
            delete_content: "Ви впевнені, що хочете видалити цей об'єкт",
            delete_title: "Видалити %{name} #%{id}",
            details: "Опис",
            error: "Під час запиту виникла помилка, і його не вдалося завершити",
            invalid_form: "Форму заповнено неправильно, будь ласка, перевірте помилки",
            loading: "Йде завантаження, будь ласка, зачекайте...",
            no: "Ні",
            not_found: "Помилка URL або ви переходите за невірним посиланням",
            yes: "Так",
            unsaved_changes:
                "Деякі з ваших змін не були збережені. Ви впевнені, що хочете їх ігнорувати?",
            bulk_update_content:
                'Ви впевнені, що хочете оновити цей %{name}? |||| Ви впевнені, що хочете оновити ці %{smart_count} елементи?',
            bulk_update_title:
                'Оновлення %{name} |||| Оновлення %{smart_count} %{name}',
        },
        navigation: {
            no_results: "Результатів не знайдено",
            no_more_results:
                "Сторінка %{page} виходить за межі нумерації, спробуйте попередню",
            page_out_of_boundaries: "Сторінка %{page} поза межами",
            page_out_from_end: "Неможливо перейти далі останньої сторінки",
            page_out_from_begin: "Номер сторінки не може бути меншим за 1",
            page_range_info: "%{offsetBegin}-%{offsetEnd} з %{total}",
            page_rows_per_page: "Рядків на сторінці:",
            next: "Наступна",
            prev: "Попередня",
            skip_nav: "Перейти до вмісту",
            partial_page_range_info:
                '%{offsetBegin}-%{offsetEnd} з більше ніж %{offsetEnd}',
            current_page: 'Сторінка %{page}',
            page: 'Перейти до сторінки %{page}',
            first: 'Перша',
            last: 'Остання',
            previous: 'Попередня',
        },
        sort: {
            sort_by: 'Сортувати за %{field} %{order}',
            ASC: 'зростанням',
            DESC: 'спаданням',
        },
        auth: {
            auth_check_error: "Будь ласка, авторизуйтесь для продовження роботи",
            user_menu: "Профіль",
            username: "Ім'я користувача",
            password: "Пароль",
            sign_in: "Увійти",
            sign_in_error: "Помилка автентифікації, спробуйте знову",
            logout: "Вийти"
        },
        notification: {
            updated: "Елемент оновлено |||| %{smart_count} оновлено |||| %{smart_count} оновлено",
            created: "Елемент створено",
            deleted: "Елемент видалено |||| %{smart_count} видалено |||| %{smart_count} видалено",
            bad_item: "Елемент не валідний",
            item_doesnt_exist: "Елемент не існує",
            http_error: "Помилка сервера",
            data_provider_error: "Помилка dataProvider, перевірте консоль",
            i18n_error: "Не вдалося завантажити переклад для вказаної мови",
            canceled: "Операцію скасовано",
            logged_out: "Ваша сесія завершена, спробуйте перепідключитись/увійти знову",
            not_authorized: "У вас немає доступу до цього ресурсу.",
        },
        validation: {
            required: "Обов'язково для заповнення",
            minLength: "Мінімальна кількість символів %{min}",
            maxLength: "Максимальна кількість символів %{max}",
            minValue: "Мінімальне значення %{min}",
            maxValue: "Значення може бути %{max} або менше",
            number: "Повинно бути числом",
            email: "Некоректний email",
            oneOf: "Повинно бути одним з: %{options}",
            regex: "Повинно бути у форматі (regexp): %{pattern}"
        },
        saved_queries: {
            label: 'Збережені запити',
            query_name: 'Назва запиту',
            new_label: 'Зберегти поточний запит...',
            new_dialog_title: 'Зберегти поточний запит як...',
            remove_label: 'Видалити запит',
            remove_label_with_name: 'Видалити запит "%{name}"',
            remove_dialog_title: 'Видалення збережених запитів?',
            remove_message:
                'Ви впевнені, що хочете видалити цей елемент зі списку збережених запитів?',
            help: 'Відфільтруйте список і збережіть цей запит',
        },
    }
};

export default ukraineMessages;
