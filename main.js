// temporary until part of the api
const weekNames = {
    "9 – 15 Oct 2023": "Week 3",
    "16 – 22 Oct 2023": "Week 4",
    "23 – 29 Oct 2023": "Week 5",
    "30 Oct – 5 Nov 2023": "Consolidation Week",
    "6 – 12 Nov 2023": "Week 6",
    "13 – 19 Nov 2023": "Week 7",
    "20 – 26 Nov 2023": "Week 8",
    "27 Nov – 3 Dec 2023": "Week 9",
    "4 – 10 Dec 2023": "Week 10",
    "11 – 17 Dec 2023": "Week 11"
};

const userCanCreateUnnamedEvents = async () => {
    return await fetch("/canCreateUnnamedEvents", { credentials: "include" }).then(r => r.json());
}

/***
 * Clicking on an Event
 */
const eventClick = (info) => {
    document.getElementById('eventTitleView').innerText = info.event.title;
    document.getElementById('eventStartTimeView').textContent = info.event.start;
    document.getElementById('eventEndTimeView').textContent = info.event.end;

    fetch("/canDelete/" + info.event.id, { credentials: "include" }).then(r => r.json()).then(allowedToDelete => {
        const deleteButton = document.getElementById('deleteEvent');
        if (allowedToDelete) {
            deleteButton.style.display = 'block';
            deleteButton.onclick = () => {
                if (confirm("Confirm deleting this booking")) {
                    fetch(`/delete/${info.event.id}`, { method: "DELETE" }).then(() => {
                        window.location.reload();
                    })
                }
            }
        } else {
            deleteButton.style.display = 'none';
        }

        $('#eventDetailsModal').modal('show');
    })
};

/**
 * Loading the Calendar
 */
document.addEventListener('DOMContentLoaded', function () {
    let calendar = new FullCalendar.Calendar(document.getElementById('calendar'), {
        headerToolbar: {
            left: '',
            center: 'title',
            right: 'prev,next today',
        },
        navLinks: false,
        nowIndicator: true,
        editable: true,
        initialView: 'timeGridWeek',
        locale: "en-GB",
        firstDay: 1,
        allDaySlot: false,
        eventClick: eventClick,
        events: "/get",
        selectable: true,
        selectMirror: true,
        selectOverlap: false,
        select: (info) => {
            console.log(info.start)
            console.log(info.start.toISOString().slice(0, 19))
            const formatDate = d => `${d.getFullYear()}-${(d.getMonth() + 1).toString().padStart(2, "0")}-${d.getDate().toString().padStart(2, "0")}T${d.getHours().toString().padStart(2, "0")}:${d.getMinutes().toString().padStart(2, "0")}:00`;
            console.log(formatDate(info.start))
            document.getElementById("eventStartTime").value = formatDate(info.start);
            document.getElementById("eventEndTime").value = formatDate(info.end);
            document.getElementById("create-button").click();
        }
    });
    calendar.render();

    if (window.innerWidth < 1000) {
        calendar.changeView("listWeek");
    }

    // Week Names
    document.getElementById("week-name").innerText = weekNames[document.getElementById("fc-dom-1").innerText] || "";
    document.getElementById("fc-dom-1").addEventListener("DOMCharacterDataModified", function () {
        document.getElementById("week-name").innerText = weekNames[document.getElementById("fc-dom-1").innerText] || "";
    }, false);
});

document.getElementById("create-button").onclick = async () => {
    const titleGroup = document.getElementById('titleGroup');
    titleGroup.style.display = ['Engineering', 'Meeting', 'Other'].includes(document.getElementById("eventType").value) ? 'block' : 'none';
    const nameSelectButton = document.getElementById("name-selector");
    nameSelectButton.style.display = document.getElementById("eventType").value == "Other" && await userCanCreateUnnamedEvents() ? "block" : "none";
    document.getElementById("create-error").innerText = "";
}

// Show or hide the event title input based on event type
document.getElementById('eventType').addEventListener('change', async function () {
    const selectedType = this.value;
    const titleGroup = document.getElementById('titleGroup');
    titleGroup.style.display = ['Engineering', 'Meeting', 'Other'].includes(selectedType) ? 'block' : 'none';
    const nameSelectButton = document.getElementById("name-selector");
    nameSelectButton.style.display = selectedType == "Other" && await userCanCreateUnnamedEvents() ? "block" : "none";
});

/**
 * Add an Event
 */
document.getElementById('submitEvent').addEventListener('click', async function () {
    const eventType = document.getElementById('eventType').value;
    const eventTitle = document.getElementById('eventTitle').value;
    const eventStartTime = document.getElementById('eventStartTime').value;
    const eventEndTime = document.getElementById('eventEndTime').value;
    const eventUnnamed = document.getElementById("name-selector-check").checked;

    // Make an API request to submit the event
    try {
        const response = await axios.post('/create', {
            type: eventType,
            title: eventTitle,
            start: eventStartTime,
            end: eventEndTime,
            noNameAttached: eventUnnamed,
        });

        // Handle the API response (e.g., show a success message)
        if (response.data.status === 'OK') {
            $('#eventModal').modal('hide');
            window.location.reload();
        } else {
            document.getElementById("create-error").innerText = 'Failed to create the event. Please try again or correct the data. ' + response.data;
        }
    } catch (error) {
        document.getElementById("create-error").innerText = 'An error occurred while creating the event. ' + error.response.data;
    }
});

/**
 * Page Info
 */
const eventTypeDropdown = document.getElementById('eventType');

fetch("/info", { credentials: "include" }).then(r => r.json()).then(d => {
    document.getElementById("user-logged-in").innerText = d.Name;
    document.getElementById("commit-hash").innerText = d.CommitHash;

    createTypes = [...new Set(d.CreateTypes)];
    createTypes.forEach((e) => {
        const option = document.createElement('option');
        option.value = e;
        option.text = e;
        eventTypeDropdown.appendChild(option);
    })
})
