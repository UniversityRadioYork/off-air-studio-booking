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
    "11 – 17 Dec 2023": "Week 11",
    "8 – 14 Jan 2024": "Revision Week",
    "15 – 21 Jan 2024": "R+A Week 1",
    "22 – 28 Jan 2024": "R+A Week 2",
    "29 Jan – 4 Feb 2024": "R+A Week 3",
    "5 – 11 Feb 2024": "Refreshers Week",
    "12 – 18 Feb 2024": "Week 1",
    "19 – 25 Feb 2024": "Week 2",
    "26 Feb – 3 Mar 2024": "Week 3",
    "4 – 10 Mar 2024": "Week 4",
    "11 – 17 Mar 2024": "Week 5",
    "18 – 24 Mar 2024": "Week 6"
};

let userCanCreateUnnamedEvents = false;

/***
 * Clicking on an Event
 */
const eventClick = (info) => {
    document.getElementById('eventTitleView').innerText = info.event.title;
    document.getElementById('eventStartTimeView').textContent = info.event.start;
    document.getElementById('eventEndTimeView').textContent = info.event.end;

    fetch("/canModify/" + info.event.id, { credentials: "include" }).then(r => r.json()).then(d => {
        let allowedToDelete = d.Delete;
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

        let canClaimForStation = d.ClaimForStation;
        const claimButton = document.getElementById("claimEvent");
        if (canClaimForStation) {
            claimButton.style.display = 'block';
            claimButton.onclick = () => {
                fetch(`/claim/${info.event.id}`, { method: "PUT" }).then(() => {
                    window.location.reload();
                })
            }
        } else {
            claimButton.style.display = 'none';
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
            left: 'title',
            center: '',
            right: 'prev,next today',
        },
        navLinks: false,
        nowIndicator: true,
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

    if (window.innerWidth < 750) {
        calendar.changeView("timeGridDay");
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
    nameSelectButton.style.display = document.getElementById("eventType").value == "Other" && userCanCreateUnnamedEvents ? "block" : "none";
    const repeatSelector = document.getElementById("repeat");
    repeatSelector.style.display = document.getElementById("eventType").value == "Meeting" ? "block" : "none";
    document.getElementById("create-error").innerText = "";
}

// Show or hide the event title input based on event type
document.getElementById('eventType').addEventListener('change', async function () {
    const selectedType = this.value;
    const titleGroup = document.getElementById('titleGroup');
    titleGroup.style.display = ['Engineering', 'Meeting', 'Other'].includes(selectedType) ? 'block' : 'none';
    const nameSelectButton = document.getElementById("name-selector");
    nameSelectButton.style.display = selectedType == "Other" && userCanCreateUnnamedEvents ? "block" : "none";
    const repeatSelector = document.getElementById("repeat");
    repeatSelector.style.display = selectedType == "Meeting" ? "block" : "none";
    document.getElementById("repeatEvent").value = 1;
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
    const repeatNum = Number(document.getElementById("repeatEvent").value);

    // Make an API request to submit the event
    try {
        const response = await axios.post('/create', {
            type: eventType,
            title: eventTitle,
            start: eventStartTime,
            end: eventEndTime,
            noNameAttached: eventUnnamed,
            repeat: repeatNum
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
    userCanCreateUnnamedEvents = d.UserCanCreateUnnamedEvents;

    createTypes = [...new Set(d.CreateTypes)];
    createTypes.forEach((e) => {
        const option = document.createElement('option');
        option.value = e;
        option.text = e;
        eventTypeDropdown.appendChild(option);
    })
})

$('#welcomemodal').modal('show');
